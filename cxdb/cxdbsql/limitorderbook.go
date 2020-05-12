package cxdbsql

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/mit-dci/lit/crypto/koblitz"
	"github.com/mit-dci/opencx/match"
)

// SQLLimitOrderbook is the representation of a limit orderbook for SQL
type SQLLimitOrderbook struct {
	DBHandler *sql.DB

	// db username and password
	dbUsername string
	dbPassword string

	// db host and port
	dbAddr net.Addr

	// orderbook schema name
	orderSchema string

	// this pair
	pair *match.Pair
}

// The schema for the limit orderbook
const (
	limitOrderbookSchema = "pubkey VARBINARY(66), orderID VARBINARY(64), side TEXT, price DOUBLE(30,2) UNSIGNED, amountHave BIGINT(64), amountWant BIGINT(64), time TIMESTAMP"
)

// CreateLimitOrderbook creates a limit orderbook based on a pair
func CreateLimitOrderbook(pair *match.Pair) (book match.LimitOrderbook, err error) {

	conf := new(dbsqlConfig)
	*conf = *defaultConf

	// set the default conf
	dbConfigSetup(conf)

	// Resolve new address
	var addr net.Addr
	if addr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(conf.DBHost, fmt.Sprintf("%d", conf.DBPort))); err != nil {
		err = fmt.Errorf("Couldn't resolve db address for CreateLimitEngine: %s", err)
		return
	}

	// Set values for limit engine
	lo := &SQLLimitOrderbook{
		dbUsername:  conf.DBUsername,
		dbPassword:  conf.DBPassword,
		orderSchema: conf.ReadOnlyOrderSchemaName,
		dbAddr:      addr,
		pair:        pair,
	}

	if err = lo.setupLimitOrderbookTables(); err != nil {
		err = fmt.Errorf("Error setting up limit orderbook tables while creating engine: %s", err)
		return
	}

	// Now connect to the database and create the schemas / tables
	openString := fmt.Sprintf("%s:%s@%s(%s)/", lo.dbUsername, lo.dbPassword, lo.dbAddr.Network(), lo.dbAddr.String())
	if lo.DBHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for CreateLimitEngine: %s", err)
		return
	}

	// Make sure we can actually connect
	if err = lo.DBHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: %s", err)
		return
	}

	// Actually set the return
	book = lo
	return
}

// setupLimitOrderbookTables sets up the tables needed for the limit orderbook.
// This assumes everything else is set
func (lo *SQLLimitOrderbook) setupLimitOrderbookTables() (err error) {

	openString := fmt.Sprintf("%s:%s@%s(%s)/", lo.dbUsername, lo.dbPassword, lo.dbAddr.Network(), lo.dbAddr.String())
	var rootHandler *sql.DB
	if rootHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for setup limit tables: %s", err)
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
		err = fmt.Errorf("Error when beginning transaction for setup limit tables: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while matching setup limit tables: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// Now create the schema
	if _, err = tx.Exec("CREATE SCHEMA IF NOT EXISTS " + lo.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Error creating schema for setup limit order tables: %s", err)
		return
	}

	// use the schema
	if _, err = tx.Exec("USE " + lo.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: %s", lo.orderSchema, err)
		return
	}

	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", lo.pair.String(), limitEngineSchema)
	if _, err = tx.Exec(createTableQuery); err != nil {
		err = fmt.Errorf("Error creating limit orderbook table: %s", err)
		return
	}
	return
}

// UpdateBookExec takes in an order execution and updates the orderbook.
func (lo *SQLLimitOrderbook) UpdateBookExec(orderExec *match.OrderExecution) (err error) {
	// doing a bunch of stuff in a tx because ACID
	var tx *sql.Tx
	if tx, err = lo.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for UpdateBookExec: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while running UpdateBookExec: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// First use the limit schema
	if _, err = tx.Exec("USE " + lo.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using limit schema to UpdateBookExec: %s", err)
		return
	}

	// If the order was filled then delete it. If not then update it.
	if orderExec.Filled {
		// If the order was filled, delete it from the orderbook
		deleteOrderQuery := fmt.Sprintf("DELETE FROM %s WHERE orderID='%x';", lo.pair.String(), orderExec.OrderID)
		// var res sql.Result
		if _, err = tx.Exec(deleteOrderQuery); err != nil {
			err = fmt.Errorf("Error deleting order within tx for UpdateBookExec: %s", err)
			return
		}
		// Now we check that there was only one row deleted. If there were more then we log it and move on. Shouldn't have put those orders there in the first place.
		// var rowsAffected int64
		// if rowsAffected, err = res.RowsAffected(); err != nil {
		// 	err = fmt.Errorf("Error while getting rows affected for UpdateBookExec: %s", err)
		// 	return
		// }
		// if rowsAffected != 1 {
		// 	err = fmt.Errorf("Error: Order delete should only have affected one row. Instead, it affected %d", rowsAffected)
		// 	return
		// }
	} else {
		// If the order was not filled, just update the amounts
		updateOrderQuery := fmt.Sprintf("UPDATE %s SET amountHave=%d, amountWant=%d WHERE orderID='%x';", lo.pair.String(), orderExec.NewAmountHave, orderExec.NewAmountWant, orderExec.OrderID)
		// var res sql.Result
		if _, err = tx.Exec(updateOrderQuery); err != nil {
			err = fmt.Errorf("Error updating order within tx for UpdateBookExec: %s", err)
			return
		}

		// Now we check that there was only one row updated. If there were more then we log it and move on. Shouldn't have put those orders there in the first place.
		// var rowsAffected int64
		// if rowsAffected, err = res.RowsAffected(); err != nil {
		// 	err = fmt.Errorf("Error while getting rows affected for UpdateBookExec: %s", err)
		// 	return
		// }
		// if rowsAffected != 1 {
		// 	logging.Errorf("Error: Order update should only have affected one row. Instead, it affected %d", rowsAffected)
		// }
	}
	return
}

// UpdateBookCancel takes in an order cancellation and updates the orderbook.
func (lo *SQLLimitOrderbook) UpdateBookCancel(cancel *match.CancelledOrder) (err error) {
	// Transaction so we're acid
	var tx *sql.Tx
	if tx, err = lo.DBHandler.Begin(); err != nil {
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

	// First use the limit schema
	if _, err = tx.Exec("USE " + lo.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using limit schema for UpdateBookCancel: %s", err)
		return
	}

	// The order was filled, delete it from the orderbook
	deleteOrderQuery := fmt.Sprintf("DELETE FROM %s WHERE orderID='%x';", lo.pair.String(), cancel.OrderID)
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

// UpdateBookPlace takes in an order, ID, timestamp, and adds the order to the orderbook.
func (lo *SQLLimitOrderbook) UpdateBookPlace(limitIDPair *match.LimitOrderIDPair) (err error) {
	// ACID
	var tx *sql.Tx
	if tx, err = lo.DBHandler.Begin(); err != nil {
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

	// First use the order schema
	if _, err = tx.Exec("USE " + lo.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using order schema for UpdateBookPlace: %s", err)
		return
	}

	insertOrderQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', '%x', '%s', %f, %d, %d, '%s');", lo.pair.String(), limitIDPair.Order.Pubkey, limitIDPair.OrderID[:], limitIDPair.Order.Side.String(), limitIDPair.Price, limitIDPair.Order.AmountHave, limitIDPair.Order.AmountWant, limitIDPair.Timestamp.Format(sqlTimeFormat))
	if _, err = tx.Exec(insertOrderQuery); err != nil {
		err = fmt.Errorf("Error placing order into db for UpdateBookPlace: %s", err)
		return
	}

	// It's just a simple insert so we're done
	return
}

// GetOrder gets an order from an OrderID
func (lo *SQLLimitOrderbook) GetOrder(orderID *match.OrderID) (limOrder *match.LimitOrderIDPair, err error) {
	limOrder = new(match.LimitOrderIDPair)
	limOrder.Order = new(match.LimitOrder)
	// Transaction so we're acid
	var tx *sql.Tx
	if tx, err = lo.DBHandler.Begin(); err != nil {
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

	if _, err = tx.Exec("USE " + lo.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using order schema for GetOrdersForPubkey: %s", err)
		return
	}

	var row *sql.Row
	getOrdersQuery := fmt.Sprintf("SELECT pubkey, side, price, orderID, amountHave, amountWant, time FROM %s WHERE orderID='%x';", lo.pair.String(), orderID[:])
	row = tx.QueryRow(getOrdersQuery)

	// we create these here so we don't take up a ton of memory allocating space for new intermediate arrays
	var pkBytes []byte
	var hashedOrderBytes []byte
	var thisPrice float64
	var sideString string
	var timeString string
	// scan the things we can into this order
	if err = row.Scan(&pkBytes, &sideString, &thisPrice, &hashedOrderBytes, &limOrder.Order.AmountHave, &limOrder.Order.AmountWant, &timeString); err != nil {
		err = fmt.Errorf("Error scanning into order for GetOrdersForPubkey: %s", err)
		return
	}

	if limOrder.Timestamp, err = time.Parse(sqlTimeFormat, timeString); err != nil {
		err = fmt.Errorf("Error parsing timestamp from string for GetOrdersForPubkey: %s", err)
		return
	}

	var sideReceiver *match.Side = new(match.Side)
	if err = sideReceiver.FromString(sideString); err != nil {
		err = fmt.Errorf("Error getting side from string for GetOrdersForPubkey: %s", err)
		return
	}

	// decode them all weirdly because of the way mysql may store the bytes
	if pkBytes, err = hex.DecodeString(string(pkBytes)); err != nil {
		err = fmt.Errorf("Error decoding bytes for GetOrdersForPubkey: %s", err)
		return
	}

	// we prepared again
	if err = limOrder.OrderID.UnmarshalText(hashedOrderBytes); err != nil {
		err = fmt.Errorf("Error unmarshalling order ID for GetOrdersForPubkey: %s", err)
		return
	}

	// Copy all of the bytes and values
	copy(limOrder.Order.Pubkey[:], pkBytes)
	limOrder.Order.TradingPair = *lo.pair
	limOrder.Order.Side = *sideReceiver
	limOrder.Price = thisPrice
	return
}

// CalculatePrice takes in a pair and returns the calculated price based on the orderbook. This is based on the midpoint of the spread.
func (lo *SQLLimitOrderbook) CalculatePrice() (price float64, err error) {
	// Transaction so we're acid
	var tx *sql.Tx
	if tx, err = lo.DBHandler.Begin(); err != nil {
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

	// First use the order schema
	if _, err = tx.Exec("USE " + lo.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using order schema for limit CalculatePrice: %s", err)
		return
	}

	sellSide := new(match.Side)
	buySide := new(match.Side)
	*sellSide = match.Sell
	*buySide = match.Buy
	// First get the max buy price and max sell price
	var maxSellRow *sql.Row
	getMaxSellPrice := fmt.Sprintf("SELECT MAX(price) FROM %s WHERE side='%s';", lo.pair.String(), sellSide.String())
	// errors for queryrow are deferred until scan -- this is important, that's why we don't err != nil here
	maxSellRow = tx.QueryRow(getMaxSellPrice)

	var maxSell float64
	if err = maxSellRow.Scan(&maxSell); err != nil {
		err = fmt.Errorf("Error scanning max sell row for limit CalculatePrice: %s", err)
		return
	}

	var minBuyRow *sql.Row
	getminBuyPrice := fmt.Sprintf("SELECT MIN(price) FROM %s WHERE side='%s';", lo.pair.String(), buySide.String())
	// errors for queryrow are deferred until scan -- this is important, that's why we don't err != nil here
	minBuyRow = tx.QueryRow(getminBuyPrice)

	var minBuy float64
	if err = minBuyRow.Scan(&minBuy); err != nil {
		err = fmt.Errorf("Error scanning min buy row for limit CalculatePrice: %s", err)
		return
	}

	price = (minBuy + maxSell) / 2
	return
}

// GetOrdersForPubkey gets orders for a specific pubkey.
func (lo *SQLLimitOrderbook) GetOrdersForPubkey(pubkey *koblitz.PublicKey) (orders map[float64][]*match.LimitOrderIDPair, err error) {
	// Make the book!!!!
	orders = make(map[float64][]*match.LimitOrderIDPair)

	// Transaction so we're acid
	var tx *sql.Tx
	if tx, err = lo.DBHandler.Begin(); err != nil {
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

	if _, err = tx.Exec("USE " + lo.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using order schema for GetOrdersForPubkey: %s", err)
		return
	}

	var rows *sql.Rows
	getOrdersQuery := fmt.Sprintf("SELECT pubkey, side, price, orderID, amountHave, amountWant, time FROM %s WHERE pubkey='%x';", lo.pair.String(), pubkey.SerializeCompressed())
	if rows, err = tx.Query(getOrdersQuery); err != nil {
		err = fmt.Errorf("Error querying for sell orders for GetOrdersForPubkey: %s", err)
		return
	}

	// we allocate space for new orders but only need one pointer
	var thisOrder *match.LimitOrder
	var thisOrderPair *match.LimitOrderIDPair

	// we create these here so we don't take up a ton of memory allocating space for new intermediate arrays
	var pkBytes []byte
	var hashedOrderBytes []byte
	var thisPrice float64
	var sideString string
	var timeString string
	for rows.Next() {
		// scan the things we can into this order
		thisOrder = new(match.LimitOrder)
		thisOrderPair = new(match.LimitOrderIDPair)
		if err = rows.Scan(&pkBytes, &sideString, &thisPrice, &hashedOrderBytes, &thisOrder.AmountHave, &thisOrder.AmountWant, &timeString); err != nil {
			err = fmt.Errorf("Error scanning into order for GetOrdersForPubkey: %s", err)
			return
		}

		if thisOrderPair.Timestamp, err = time.Parse(sqlTimeFormat, timeString); err != nil {
			err = fmt.Errorf("Error parsing timestamp for GetOrdersForPubkey: %s", err)
			return
		}

		var sideReceiver *match.Side = new(match.Side)
		if err = sideReceiver.FromString(sideString); err != nil {
			err = fmt.Errorf("Error getting side from string for GetOrdersForPubkey: %s", err)
			return
		}

		// decode them all weirdly because of the way mysql may store the bytes
		if pkBytes, err = hex.DecodeString(string(pkBytes)); err != nil {
			err = fmt.Errorf("Error decoding bytes for GetOrdersForPubkey: %s", err)
			return
		}

		// we prepared again
		if err = thisOrderPair.OrderID.UnmarshalText(hashedOrderBytes); err != nil {
			err = fmt.Errorf("Error unmarshalling order ID for GetOrdersForPubkey: %s", err)
			return
		}

		// Copy all of the bytes and values
		copy(thisOrder.Pubkey[:], pkBytes)
		thisOrder.TradingPair = *lo.pair
		thisOrderPair.Order = thisOrder
		thisOrderPair.Order.Side = *sideReceiver
		thisOrderPair.Price = thisPrice
		orders[thisPrice] = append(orders[thisPrice], thisOrderPair)

	}

	if err = rows.Close(); err != nil {
		err = fmt.Errorf("Error closing rows for GetOrdersForPubkey: %s", err)
		return
	}
	return
}

// ViewLimitOrderbook takes in a trading pair and returns the orderbook as a map
func (lo *SQLLimitOrderbook) ViewLimitOrderBook() (book map[float64][]*match.LimitOrderIDPair, err error) {
	// Make the book!!!!
	book = make(map[float64][]*match.LimitOrderIDPair)

	// Transaction so we're acid
	var tx *sql.Tx
	if tx, err = lo.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for ViewOrderBook: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error with ViewOrderBook: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec("USE " + lo.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using order schema for ViewOrderBook: %s", err)
		return
	}

	var rows *sql.Rows
	getOrdersQuery := fmt.Sprintf("SELECT pubkey, side, price, orderID, amountHave, amountWant, time FROM %s;", lo.pair.String())
	if rows, err = tx.Query(getOrdersQuery); err != nil {
		err = fmt.Errorf("Error querying for sell orders for ViewOrderBook: %s", err)
		return
	}

	// we allocate space for new orders but only need one pointer
	var thisOrder *match.LimitOrder
	var thisOrderPair *match.LimitOrderIDPair

	// we create these here so we don't take up a ton of memory allocating space for new intermediate arrays
	var pkBytes []byte
	var hashedOrderBytes []byte
	var thisPrice float64
	var sideString string
	var timeString string
	for rows.Next() {
		// scan the things we can into this order
		thisOrder = new(match.LimitOrder)
		thisOrderPair = new(match.LimitOrderIDPair)
		if err = rows.Scan(&pkBytes, &sideString, &thisPrice, &hashedOrderBytes, &thisOrder.AmountHave, &thisOrder.AmountWant, &timeString); err != nil {
			err = fmt.Errorf("Error scanning into order for ViewOrderBook: %s", err)
			return
		}

		if thisOrderPair.Timestamp, err = time.Parse(sqlTimeFormat, timeString); err != nil {
			err = fmt.Errorf("Error parsing timestamp from db for ViewOrderBook: %s", err)
			return
		}

		var sideReceiver *match.Side = new(match.Side)
		if err = sideReceiver.FromString(sideString); err != nil {
			err = fmt.Errorf("Error getting side from string for ViewOrderBook: %s", err)
			return
		}

		// decode them all weirdly because of the way mysql may store the bytes
		if pkBytes, err = hex.DecodeString(string(pkBytes)); err != nil {
			err = fmt.Errorf("Error decoding bytes for ViewOrderBook: %s", err)
			return
		}

		// we prepared again
		if err = thisOrderPair.OrderID.UnmarshalText(hashedOrderBytes); err != nil {
			err = fmt.Errorf("Error unmarshalling order ID for ViewOrderBook: %s", err)
			return
		}

		// Copy all of the bytes and values
		copy(thisOrder.Pubkey[:], pkBytes)
		thisOrder.TradingPair = *lo.pair
		thisOrderPair.Order = thisOrder
		thisOrderPair.Order.Side = *sideReceiver
		thisOrderPair.Price = thisPrice
		book[thisPrice] = append(book[thisPrice], thisOrderPair)

	}

	if err = rows.Close(); err != nil {
		err = fmt.Errorf("Error closing rows for ViewOrderBook: %s", err)
		return
	}
	return
}

// CreateLimitOrderbookMap creates a map of pair to deposit store, given a list of pairs.
func CreateLimitOrderbookMap(pairList []*match.Pair) (orderbookMap map[match.Pair]match.LimitOrderbook, err error) {

	orderbookMap = make(map[match.Pair]match.LimitOrderbook)
	var curLimOrderbook match.LimitOrderbook
	for _, pair := range pairList {
		if curLimOrderbook, err = CreateLimitOrderbook(pair); err != nil {
			err = fmt.Errorf("Error creating single limit engine while creating limit engine map: %s", err)
			return
		}
		orderbookMap[*pair] = curLimOrderbook
	}

	return
}
