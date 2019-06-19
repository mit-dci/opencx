package cxdbsql

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
)

// SQLAuctionOrderbook is the representation of a auction orderbook for SQL
type SQLAuctionOrderbook struct {
	DBHandler *sql.DB

	// db username and password
	dbUsername string
	dbPassword string

	// db host and port
	dbAddr net.Addr

	// orderbook schema name
	auctionOrderSchema string

	// this pair
	pair *match.Pair
}

// The schema for the auction orderbook
const (
	auctionOrderbookSchema = "pubkey VARBINARY(66), orderID VARBINARY(64), side TEXT, price DOUBLE(30,2) UNSIGNED, amountHave BIGINT(64), amountWant BIGINT(64), time TIMESTAMP"
)

// CreateAuctionOrderbook creates a auction orderbook based on a pair
func CreateAuctionOrderbook(pair *match.Pair) (book match.AuctionOrderbook, err error) {

	conf := new(dbsqlConfig)
	*conf = *defaultConf

	// set the default conf
	dbConfigSetup(conf)

	// Resolve new address
	var addr net.Addr
	if addr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(conf.DBHost, fmt.Sprintf("%d", conf.DBPort))); err != nil {
		err = fmt.Errorf("Couldn't resolve db address for CreateAuctionEngine: %s", err)
		return
	}

	// Set values for auction engine
	ao := &SQLAuctionOrderbook{
		dbUsername:         conf.DBUsername,
		dbPassword:         conf.DBPassword,
		auctionOrderSchema: conf.ReadOnlyAuctionSchemaName,
		dbAddr:             addr,
		pair:               pair,
	}

	if err = ao.setupAuctionOrderbookTables(); err != nil {
		err = fmt.Errorf("Error setting up auction orderbook tables while creating engine: %s", err)
		return
	}

	// Now connect to the database and create the schemas / tables
	openString := fmt.Sprintf("%s:%s@%s(%s)/", ao.dbUsername, ao.dbPassword, ao.dbAddr.Network(), ao.dbAddr.String())
	if ao.DBHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for CreateAuctionEngine: %s", err)
		return
	}

	// Make sure we can actually connect
	if err = ao.DBHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: %s", err)
		return
	}

	return
}

// setupAuctionOrderbookTables sets up the tables needed for the auction orderbook.
// This assumes everything else is set
func (ao *SQLAuctionOrderbook) setupAuctionOrderbookTables() (err error) {

	openString := fmt.Sprintf("%s:%s@%s(%s)/", ao.dbUsername, ao.dbPassword, ao.dbAddr.Network(), ao.dbAddr.String())
	var rootHandler *sql.DB
	if rootHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for setup auction orderbook tables: %s", err)
		return
	}

	// when we're done close please
	defer rootHandler.Close()

	if err = rootHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: %s", err)
		return
	}

	// We do this in a transaction because it's more than one operation
	var tx *sql.Tx
	if tx, err = rootHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for setup auction orderbook tables: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while setting up auction orderbook tables: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// Now create the schema
	if _, err = tx.Exec("CREATE SCHEMA IF NOT EXISTS " + ao.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Error creating schema for setup auction order tables: %s", err)
		return
	}

	// use the schema
	if _, err = tx.Exec("USE " + ao.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: %s", ao.auctionOrderSchema, err)
		return
	}

	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", ao.pair.String(), auctionEngineSchema)
	if _, err = tx.Exec(createTableQuery); err != nil {
		err = fmt.Errorf("Error creating auction orderbook table: %s", err)
		return
	}
	return
}

// UpdateBookExec takes in an order execution and updates the orderbook.
func (ao *SQLAuctionOrderbook) UpdateBookExec(exec *match.OrderExecution) (err error) {
	// We do this in a transaction because it's more than one operation
	var tx *sql.Tx
	if tx, err = ao.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for setup auction orderbook tables: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while setting up auction orderbook tables: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// First use the auction schema
	if _, err = tx.Exec("USE " + ao.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using auction schema to process order execution: %s", err)
		return
	}

	// If the order was filled then delete it. If not then update it.
	if exec.Filled {
		// If the order was filled, delete it from the orderbook
		deleteOrderQuery := fmt.Sprintf("DELETE FROM %s WHERE hashedOrder='%x';", ao.pair.String(), exec.OrderID)
		var res sql.Result
		if res, err = tx.Exec(deleteOrderQuery); err != nil {
			err = fmt.Errorf("Error deleting order within tx for processorderexecution: %s", err)
			return
		}
		// Now we check that there was only one row deleted. If there were more then we log it and move on. Shouldn't have put those orders there in the first place.
		var rowsAffected int64
		if rowsAffected, err = res.RowsAffected(); err != nil {
			err = fmt.Errorf("Error while getting rows affected for process order execution: %s", err)
			return
		}
		if rowsAffected != 1 {
			logging.Errorf("Error: Order delete should only have affected one row. Instead, it affected %d", rowsAffected)
		}
	} else {
		// If the order was not filled, just update the amounts
		updateOrderQuery := fmt.Sprintf("UPDATE %s SET amountHave=%d, amountWant=%d WHERE hashedOrder='%x';", ao.pair.String(), exec.NewAmountHave, exec.NewAmountWant, exec.OrderID)
		var res sql.Result
		if res, err = tx.Exec(updateOrderQuery); err != nil {
			err = fmt.Errorf("Error updating order within tx for processorderexecution: %s", err)
			return
		}

		// Now we check that there was only one row updated. If there were more then we log it and move on. Shouldn't have put those orders there in the first place.
		var rowsAffected int64
		if rowsAffected, err = res.RowsAffected(); err != nil {
			err = fmt.Errorf("Error while getting rows affected for process order execution: %s", err)
			return
		}
		if rowsAffected != 1 {
			logging.Errorf("Error: Order update should only have affected one row. Instead, it affected %d", rowsAffected)
		}
	}
	return
}

// UpdateBookCancel takes in an order cancellation and updates the orderbook.
func (ao *SQLAuctionOrderbook) UpdateBookCancel(cancel *match.CancelledOrder) (err error) {
	// Transaction so we're acid
	var tx *sql.Tx
	if tx, err = ao.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for UpdateBookCancel: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error with UpdateBookCancel: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// First use the auction schema
	if _, err = tx.Exec("USE " + ao.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using auction schema for UpdateBookCancel: %s", err)
		return
	}

	// The order was filled, delete it from the orderbook
	deleteOrderQuery := fmt.Sprintf("DELETE FROM %s WHERE hashedOrder='%x';", ao.pair.String(), cancel.OrderID)
	var res sql.Result
	if res, err = tx.Exec(deleteOrderQuery); err != nil {
		err = fmt.Errorf("Error deleting order within tx for cancel: %s", err)
		return
	}
	// Now we check that there was only one row deleted. If there were more then we log it and move on. Shouldn't have put those orders there in the first place.
	var rowsAffected int64
	if rowsAffected, err = res.RowsAffected(); err != nil {
		err = fmt.Errorf("Error while getting rows affected for cancel: %s", err)
		return
	}
	if rowsAffected != 1 {
		err = fmt.Errorf("Error: Order cancel should only have affected one row. Instead, it affected %d", rowsAffected)
		return
	}
	return
}

// UpdateBookPlace takes in an order, ID, auction ID, and adds the order to the orderbook.
func (ao *SQLAuctionOrderbook) UpdateBookPlace(auctionIDPair *match.AuctionOrderIDPair) (err error) {
	// Transaction so we're acid
	var tx *sql.Tx
	if tx, err = ao.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for UpdateBookPlace: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error for UpdateBookPlace: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// First use the auction schema
	if _, err = tx.Exec("USE " + ao.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using auction schema for UpdateBookPlace: %s", err)
		return
	}

	logging.Infof("Placing order in orderbook: \n%s", auctionIDPair.Order)

	insertOrderQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', '%s', %f, %d, %d, '%x', '%x', '%x', '%x');", ao.pair.String(), auctionIDPair.Order.Pubkey, auctionIDPair.Order.Side, auctionIDPair.Price, auctionIDPair.Order.AmountHave, auctionIDPair.Order.AmountWant, auctionIDPair.Order.AuctionID, auctionIDPair.Order.Nonce, auctionIDPair.Order.Signature, auctionIDPair.OrderID)
	if _, err = tx.Exec(insertOrderQuery); err != nil {
		err = fmt.Errorf("Error placing order into db for UpdateBookPlace: %s", err)
		return
	}
	// It's just a simple insert so we're done

	return
}

// GetOrder gets an order from an OrderID
func (ao *SQLAuctionOrderbook) GetOrder(orderID *match.OrderID) (aucOrder *match.AuctionOrderIDPair, err error) {
	aucOrder = new(match.AuctionOrderIDPair)
	aucOrder.Order = new(match.AuctionOrder)

	// Transaction so we're acid
	var tx *sql.Tx
	if tx, err = ao.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for GetOrder: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error with GetOrder: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// First use the auction schema
	if _, err = tx.Exec("USE " + ao.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using auction schema for GetOrder: %s", err)
		return
	}

	// This is just a modified GetOrdersForPubkey
	var row *sql.Row
	selectOrderQuery := fmt.Sprintf("SELECT pubkey, side, price, amountHave, amountWant, auctionID, nonce, sig, hashedOrder FROM %s WHERE hashedOrder='%x';", ao.pair, orderID)
	// Remember: errors for this are deferred to scan
	row = tx.QueryRow(selectOrderQuery)

	// we create these here so we don't take up a ton of memory allocating space for new intermediate arrays
	var pkBytes []byte
	var auctionIDBytes []byte
	var nonceBytes []byte
	var sigBytes []byte
	var hashedOrderBytes []byte

	// scan the things we can into this order
	if err = row.Scan(&pkBytes, &aucOrder.Order.Side, &aucOrder.Price, &aucOrder.Order.AmountHave, &aucOrder.Order.AmountWant, &auctionIDBytes, &nonceBytes, &sigBytes, &hashedOrderBytes); err != nil {
		err = fmt.Errorf("Error scanning into order for GetOrdersForPubkey: %s", err)
		return
	}

	// decode them all weirdly because of the way mysql may store the bytes
	for _, byteArrayPtr := range []*[]byte{&pkBytes, &auctionIDBytes, &nonceBytes, &sigBytes} {
		if *byteArrayPtr, err = hex.DecodeString(string(*byteArrayPtr)); err != nil {
			err = fmt.Errorf("Error decoding bytes for GetOrdersForPubkey: %s", err)
			return
		}
	}

	// we prepared again
	if err = aucOrder.OrderID.UnmarshalText(hashedOrderBytes); err != nil {
		err = fmt.Errorf("Error unmarshalling order ID for GetOrdersForPubkey: %s", err)
		return
	}

	// Copy all of the bytes and values
	copy(aucOrder.Order.Pubkey[:], pkBytes)
	copy(aucOrder.Order.AuctionID[:], auctionIDBytes)
	copy(aucOrder.Order.Signature[:], sigBytes)
	copy(aucOrder.Order.Nonce[:], nonceBytes)
	aucOrder.Order.TradingPair = *ao.pair

	// It's a single row so it doesn't need to be closed
	return
}

// CalculatePrice takes in a pair and returns the calculated price based on the orderbook.
func (ao *SQLAuctionOrderbook) CalculatePrice(auctionID *match.AuctionID) (price float64, err error) {
	// Transaction so we're acid
	var tx *sql.Tx
	if tx, err = ao.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for CalculatePrice: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error with CalculatePrice: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// First use the auction schema
	if _, err = tx.Exec("USE " + ao.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using auction schema for CalculatePrice: %s", err)
		return
	}
	sellSide := new(match.Side)
	buySide := new(match.Side)
	*sellSide = match.Sell
	*buySide = match.Buy
	// First get the max buy price and max sell price
	var maxSellRow *sql.Row
	getMaxSellPrice := fmt.Sprintf("SELECT MAX(price) FROM %s WHERE side='%s' AND auctionID='%x';", ao.pair.String(), sellSide.String(), auctionID)
	// errors for queryrow are deferred until scan -- this is important, that's why we don't err != nil here
	maxSellRow = tx.QueryRow(getMaxSellPrice)

	var maxSell float64
	if err = maxSellRow.Scan(&maxSell); err != nil {
		err = fmt.Errorf("Error scanning max sell row for auction CalculatePrice: %s", err)
		return
	}

	var minBuyRow *sql.Row
	getminBuyPrice := fmt.Sprintf("SELECT MIN(price) FROM %s WHERE side='%s' AND auctionID='%x';", ao.pair.String(), buySide.String(), auctionID)
	// errors for queryrow are deferred until scan -- this is important, that's why we don't err != nil here
	minBuyRow = tx.QueryRow(getminBuyPrice)

	var minBuy float64
	if err = minBuyRow.Scan(&minBuy); err != nil {
		err = fmt.Errorf("Error scanning min buy row for auction CalculatePrice: %s", err)
		return
	}

	price = (minBuy + maxSell) / 2
	return
}

// GetOrdersForPubkey gets orders for a specific pubkey.
func (ao *SQLAuctionOrderbook) GetOrdersForPubkey(pubkey *koblitz.PublicKey) (orders map[float64][]*match.AuctionOrderIDPair, err error) {
	// Make the book!!!!
	orders = make(map[float64][]*match.AuctionOrderIDPair)

	// Transaction so we're acid
	var tx *sql.Tx
	if tx, err = ao.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for GetOrdersForPubkey: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error with GetOrdersForPubkey: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec("USE " + ao.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using auction schema for GetOrdersForPubkey: %s", err)
		return
	}

	// This is just a modified viewauctionorderbook
	var rows *sql.Rows
	selectOrderQuery := fmt.Sprintf("SELECT pubkey, side, price, amountHave, amountWant, auctionID, nonce, sig, hashedOrder FROM %s WHERE pubkey='%x';", ao.pair, pubkey.SerializeCompressed())
	if rows, err = tx.Query(selectOrderQuery); err != nil {
		err = fmt.Errorf("Error getting orders from db for GetOrdersForPubkey: %s", err)
		return
	}

	// we allocate space for new orders but only need one pointer
	var thisOrder *match.AuctionOrder
	var thisOrderPair *match.AuctionOrderIDPair

	// we create these here so we don't take up a ton of memory allocating space for new intermediate arrays
	var pkBytes []byte
	var auctionIDBytes []byte
	var nonceBytes []byte
	var sigBytes []byte
	var hashedOrderBytes []byte
	var thisPrice float64

	for rows.Next() {
		// scan the things we can into this order
		thisOrder = new(match.AuctionOrder)
		thisOrderPair = new(match.AuctionOrderIDPair)
		if err = rows.Scan(&pkBytes, &thisOrder.Side, &thisPrice, &thisOrder.AmountHave, &thisOrder.AmountWant, &auctionIDBytes, &nonceBytes, &sigBytes, &hashedOrderBytes); err != nil {
			err = fmt.Errorf("Error scanning into order for GetOrdersForPubkey: %s", err)
			return
		}

		// decode them all weirdly because of the way mysql may store the bytes
		for _, byteArrayPtr := range []*[]byte{&pkBytes, &auctionIDBytes, &nonceBytes, &sigBytes} {
			if *byteArrayPtr, err = hex.DecodeString(string(*byteArrayPtr)); err != nil {
				err = fmt.Errorf("Error decoding bytes for GetOrdersForPubkey: %s", err)
				return
			}
		}

		// we prepared again
		if err = thisOrderPair.OrderID.UnmarshalText(hashedOrderBytes); err != nil {
			err = fmt.Errorf("Error unmarshalling order ID for GetOrdersForPubkey: %s", err)
			return
		}

		// Copy all of the bytes and values
		copy(thisOrder.Pubkey[:], pkBytes)
		copy(thisOrder.AuctionID[:], auctionIDBytes)
		copy(thisOrder.Signature[:], sigBytes)
		copy(thisOrder.Nonce[:], nonceBytes)
		thisOrder.TradingPair = *ao.pair
		thisOrderPair.Order = thisOrder
		thisOrderPair.Price = thisPrice
		orders[thisPrice] = append(orders[thisPrice], thisOrderPair)

	}

	if err = rows.Close(); err != nil {
		err = fmt.Errorf("Error closing rows for GetOrdersForPubkey: %s", err)
		return
	}
	return
}

// ViewAuctionOrderbook takes in a trading pair and returns the orderbook as a map
func (ao *SQLAuctionOrderbook) ViewAuctionOrderBook() (book map[float64][]*match.AuctionOrderIDPair, err error) {
	// Make the book!!!!
	book = make(map[float64][]*match.AuctionOrderIDPair)

	// Transaction so we're acid
	var tx *sql.Tx
	if tx, err = ao.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for setup auction orderbook tables: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while setting up auction orderbook tables: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec("USE " + ao.auctionOrderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using auction schema for viewauctionorderbook: %s", err)
		return
	}

	var rows *sql.Rows
	selectOrderQuery := fmt.Sprintf("SELECT pubkey, side, price, amountHave, amountWant, auctionID, nonce, sig, hashedOrder FROM %s;", ao.pair)
	if rows, err = tx.Query(selectOrderQuery); err != nil {
		err = fmt.Errorf("Error getting orders from db for viewauctionorderbook: %s", err)
		return
	}

	// we allocate space for new orders but only need one pointer
	var thisOrder *match.AuctionOrder
	var thisOrderPair *match.AuctionOrderIDPair

	// we create these here so we don't take up a ton of memory allocating space for new intermediate arrays
	var pkBytes []byte
	var auctionIDBytes []byte
	var nonceBytes []byte
	var sigBytes []byte
	var hashedOrderBytes []byte
	var thisPrice float64

	for rows.Next() {
		// scan the things we can into this order
		thisOrder = new(match.AuctionOrder)
		thisOrderPair = new(match.AuctionOrderIDPair)
		if err = rows.Scan(&pkBytes, &thisOrder.Side, &thisPrice, &thisOrder.AmountHave, &thisOrder.AmountWant, &auctionIDBytes, &nonceBytes, &sigBytes, &hashedOrderBytes); err != nil {
			err = fmt.Errorf("Error scanning into order for viewauctionorderbook: %s", err)
			return
		}

		// decode them all weirdly because of the way mysql may store the bytes
		for _, byteArrayPtr := range []*[]byte{&pkBytes, &auctionIDBytes, &nonceBytes, &sigBytes} {
			if *byteArrayPtr, err = hex.DecodeString(string(*byteArrayPtr)); err != nil {
				err = fmt.Errorf("Error decoding bytes for viewauctionorderbook: %s", err)
				return
			}
		}

		// we prepared again
		if err = thisOrderPair.OrderID.UnmarshalText(hashedOrderBytes); err != nil {
			err = fmt.Errorf("Error unmarshalling order ID for ViewAuctionOrderBook: %s", err)
			return
		}

		// Copy all of the bytes and values
		copy(thisOrder.Pubkey[:], pkBytes)
		copy(thisOrder.AuctionID[:], auctionIDBytes)
		copy(thisOrder.Signature[:], sigBytes)
		copy(thisOrder.Nonce[:], nonceBytes)
		thisOrder.TradingPair = *ao.pair
		thisOrderPair.Order = thisOrder
		thisOrderPair.Price = thisPrice
		book[thisPrice] = append(book[thisPrice], thisOrderPair)

	}

	if err = rows.Close(); err != nil {
		err = fmt.Errorf("Error closing rows for viewauctionorderbook: %s", err)
		return
	}
	return
}

// CreateAuctionOrderbookMap creates a map of pair to auction engine, given a list of pairs.
func CreateAuctionOrderbookMap(pairList []*match.Pair) (aucMap map[match.Pair]match.AuctionOrderbook, err error) {

	aucMap = make(map[match.Pair]match.AuctionOrderbook)
	var curAucEng match.AuctionOrderbook
	for _, pair := range pairList {
		if curAucEng, err = CreateAuctionOrderbook(pair); err != nil {
			err = fmt.Errorf("Error creating single auction engine while creating auction engine map: %s", err)
			return
		}
		aucMap[*pair] = curAucEng
	}

	return
}
