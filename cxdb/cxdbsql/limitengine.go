package cxdbsql

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/mit-dci/opencx/logging"
	"github.com/mit-dci/opencx/match"
	"golang.org/x/crypto/sha3"
)

// SQLLimitEngine is a struct that represents a limit matching engine with SQL as a db backend
type SQLLimitEngine struct {
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

// The schema for the limit orderbook -- TODO: THE PRICE SCHEMA SHOULD BE CONFIGURED BASED ON DESIRED PRECISION, WHICH SHOULD BE ENFORCED BY OUR TYPES AS WELL
const (
	limitEngineSchema = "pubkey VARBINARY(66), orderID VARBINARY(64), side TEXT, price DOUBLE(32,16) UNSIGNED, amountHave BIGINT(64), amountWant BIGINT(64), time TIMESTAMP"
	sqlTimeFormat     = "2006-01-02 15:04:05"
)

func CreateLimEngineStructWithConf(pair *match.Pair, conf *dbsqlConfig) (engine *SQLLimitEngine, err error) {
	// Set the default conf
	dbConfigSetup(conf)

	// Resolve new address
	var addr net.Addr
	if addr, err = net.ResolveTCPAddr("tcp", net.JoinHostPort(conf.DBHost, fmt.Sprintf("%d", conf.DBPort))); err != nil {
		err = fmt.Errorf("Couldn't resolve db address for CreateLimitEngineWithConf: %s", err)
		return
	}

	// Set values
	le := &SQLLimitEngine{
		dbUsername:  conf.DBUsername,
		dbPassword:  conf.DBPassword,
		orderSchema: conf.OrderSchemaName,
		dbAddr:      addr,
		pair:        pair,
	}

	if err = le.setupLimitOrderbookTables(); err != nil {
		err = fmt.Errorf("Error setting up limit orderbook tables while creating engine: %s", err)
		return
	}

	// Now connect to the database and create the schemas / tables
	openString := fmt.Sprintf("%s:%s@%s(%s)/", le.dbUsername, le.dbPassword, le.dbAddr.Network(), le.dbAddr.String())
	if le.DBHandler, err = sql.Open("mysql", openString); err != nil {
		err = fmt.Errorf("Error opening database for CreateLimitEngineWithConf: %s", err)
		return
	}

	// Make sure we can actually connect
	if err = le.DBHandler.Ping(); err != nil {
		err = fmt.Errorf("Could not ping the database, is it running: %s", err)
		return
	}

	// now we actually set the return, all checks have passed
	engine = le
	return
}

func CreateLimitEngineWithConf(pair *match.Pair, conf *dbsqlConfig) (engine match.LimitEngine, err error) {
	var le *SQLLimitEngine
	if le, err = CreateLimEngineStructWithConf(pair, conf); err != nil {
		err = fmt.Errorf("Error creating limit engine struct with conf for CreateLimitEngineWithConf: %s", err)
		return
	}
	engine = le
	return
}

// CreateLimitEngine creates a limit matching engine that operates using SQL as a database
func CreateLimitEngine(pair *match.Pair) (engine match.LimitEngine, err error) {

	conf := new(dbsqlConfig)
	*conf = *defaultConf

	if engine, err = CreateLimitEngineWithConf(pair, conf); err != nil {
		err = fmt.Errorf("Error creating limit engine with conf for CreateLimitEngine: %s", err)
		return
	}
	return
}

// setupLimitOrderbookTables sets up the tables needed for the limit orderbook.
// This assumes everything else is set
func (le *SQLLimitEngine) setupLimitOrderbookTables() (err error) {

	openString := fmt.Sprintf("%s:%s@%s(%s)/", le.dbUsername, le.dbPassword, le.dbAddr.Network(), le.dbAddr.String())
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
	if _, err = tx.Exec("CREATE SCHEMA IF NOT EXISTS " + le.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Error creating schema for setup limit order tables: %s", err)
		return
	}

	// use the schema
	if _, err = tx.Exec("USE " + le.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Could not use %s schema: %s", le.orderSchema, err)
		return
	}

	createTableQuery := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", le.pair.String(), limitEngineSchema)
	if _, err = tx.Exec(createTableQuery); err != nil {
		err = fmt.Errorf("Error creating limit orderbook table: %s", err)
		return
	}
	return
}

// DestroyHandler closes the DB handler that we created, and makes it nil
func (le *SQLLimitEngine) DestroyHandler() (err error) {
	if le.DBHandler == nil {
		err = fmt.Errorf("Error, cannot destroy nil handler, please create new engine")
		return
	}
	if err = le.DBHandler.Close(); err != nil {
		err = fmt.Errorf("Error closing engine handler for DestroyHandler: %s", err)
		return
	}
	le.DBHandler = nil
	return
}

// PlaceLimitOrder places an order in the limit matching engine.
// This assumes that the order is valid and is for the same pair as the matching engine
func (le *SQLLimitEngine) PlaceLimitOrder(order *match.LimitOrder) (idRes *match.LimitOrderIDPair, err error) {
	if order == nil {
		err = fmt.Errorf("Cannot place nil order, please enter valid input")
		return
	}

	if le.pair == nil {
		err = fmt.Errorf("Cannot place order with nil pair, please enter valid input")
		return
	}

	if le.DBHandler == nil {
		err = fmt.Errorf("Cannot place order with nil DBHandler, please set up limit engine correctly")
		return
	}

	// First, get the time.
	placementTime := time.Now()
	placementTimeFormatted := placementTime.Format(sqlTimeFormat)

	// Do these first so we don't have to rollback any tx's if they're wrong
	// hash order so we can use that as a primary key
	hasher := sha3.New256()
	var orderBytes []byte
	if orderBytes, err = order.Serialize(); err != nil {
		err = fmt.Errorf("Error serializing while placing order: %s", err)
		return
	}
	hasher.Write(orderBytes)
	hashedOrder := hasher.Sum(nil)

	// calculate price
	var price float64
	if price, err = order.Price(); err != nil {
		err = fmt.Errorf("Error getting price from order while placing order: %s", err)
		return
	}

	// Finally, set the auction order / id pair
	loid := &match.LimitOrderIDPair{
		OrderID:   new(match.OrderID),
		Order:     order,
		Price:     price,
		Timestamp: placementTime,
	}

	if err = loid.OrderID.UnmarshalBinary(hashedOrder); err != nil {
		err = fmt.Errorf("Could not unmarshal orderdi for PlaceLimitOrder: %s", err)
		return
	}

	// TODO: JANK ALERT, converting to string and back because that's how we insert our prices,
	// by doing a janky sprintf. This is our janky way of making sure our janky insert is never broken.
	var zeroCheckFloat float64
	if zeroCheckFloat, err = strconv.ParseFloat(fmt.Sprintf("%.16f", price), 64); err != nil {
		err = fmt.Errorf("Error parsing float: %s", err)
		return
	}

	// if the price is less than a certain amount, auto-cancel the order
	if zeroCheckFloat == float64(0) {
		err = fmt.Errorf("Placing 0-valued order is not allowed")
		return
	}

	var tx *sql.Tx
	if tx, err = le.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error beginning transaction while placing order: \n%s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while placing order: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	if _, err = tx.Exec("USE " + le.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using order schema while matching limit orders: %s", err)
		return
	}

	placeOrderQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%x', '%x', '%s', %f, %d, %d, '%s');", le.pair.String(), order.Pubkey[:], hashedOrder, order.Side.String(), price, order.AmountHave, order.AmountWant, placementTimeFormatted)
	if _, err = tx.Exec(placeOrderQuery); err != nil {
		err = fmt.Errorf("Error placing order into db for PlaceLimitOrder: %s", err)
		return
	}

	idRes = loid
	return
}

// CancelLimitOrder cancels an auction order, this assumes that the auction order actually exists
func (le *SQLLimitEngine) CancelLimitOrder(orderID *match.OrderID) (cancelled *match.CancelledOrder, cancelSettlement *match.SettlementExecution, err error) {

	var tx *sql.Tx
	if tx, err = le.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for CancelLimitOrder: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error for CancelLimitOrder: \n%s", err)
			return
		}
		err = tx.Commit()
		return
	}()

	if _, err = tx.Exec("USE " + le.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using order schema while cancelling limit order: %s", err)
		return
	}

	var rows *sql.Rows
	selectOrderQuery := fmt.Sprintf("SELECT pubkey, side, amountHave FROM %s WHERE orderID = '%x' FOR UPDATE;", le.pair, orderID)
	if rows, err = tx.Query(selectOrderQuery); err != nil {
		err = fmt.Errorf("Error getting order from db for CancelLimitOrder: %s", err)
		return
	}

	actualSide := new(match.Side)

	var pkBytes []byte
	var orderSide string
	var remainingHave uint64
	if rows.Next() {
		// scan the things we can into this order
		if err = rows.Scan(&pkBytes, &orderSide, &remainingHave); err != nil {
			err = fmt.Errorf("Error scanning for order for CancelLimitOrder: %s", err)
			return
		}

		// decode them all weirdly because of the way mysql may store the bytes
		if pkBytes, err = hex.DecodeString(string(pkBytes)); err != nil {
			err = fmt.Errorf("Error decoding pkBytes for CancelLimitOrder: %s", err)
			return
		}

		if err = actualSide.FromString(orderSide); err != nil {
			err = fmt.Errorf("Error getting side from string for CancelLimitOrder: %s", err)
			return
		}

	}

	deleteOrderQuery := fmt.Sprintf("DELETE FROM %s WHERE orderID = '%x';", le.pair.String(), orderID)
	if _, err = tx.Exec(deleteOrderQuery); err != nil {
		err = fmt.Errorf("Error deleting order for CancelLimitOrder: %s", err)
		return
	}

	cancelled = &match.CancelledOrder{
		OrderID: orderID,
	}
	var debitAsset match.Asset
	if *actualSide == match.Buy {
		debitAsset = le.pair.AssetHave
	} else {
		debitAsset = le.pair.AssetWant
	}
	cancelSettlement = &match.SettlementExecution{
		Amount: remainingHave,
		Type:   match.Debit,
		Asset:  debitAsset,
	}
	copy(cancelSettlement.Pubkey[:], pkBytes)

	return
}

// MatchLimitOrders matches limit orders based on price/time priority
func (le *SQLLimitEngine) MatchLimitOrders() (orderExecs []*match.OrderExecution, settlementExecs []*match.SettlementExecution, err error) {
	if le.DBHandler == nil {
		err = fmt.Errorf("Cannot match orders for nil handler, please recreate engine")
		return
	}

	var tx *sql.Tx
	if tx, err = le.DBHandler.Begin(); err != nil {
		err = fmt.Errorf("Error when beginning transaction for MatchLimitOrders: %s", err)
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error for MatchLimitOrders: \n%s", err)
			return
		}
		err = tx.Commit()
		return
	}()

	if _, err = tx.Exec("USE " + le.orderSchema + ";"); err != nil {
		err = fmt.Errorf("Error using order schema while matching limit orders: %s", err)
		return
	}

	sellSide := new(match.Side)
	buySide := new(match.Side)
	*sellSide = match.Sell
	*buySide = match.Buy
	// First get the max buy price and max sell price
	var maxSellRow *sql.Row
	getMaxSellPrice := fmt.Sprintf("SELECT MAX(price) FROM %s WHERE side='%s' FOR UPDATE;", le.pair.String(), sellSide.String())
	// errors for queryrow are deferred until scan -- this is important, that's why we don't err != nil here
	logging.Infof("maxsell: %s", sellSide.String())
	maxSellRow = tx.QueryRow(getMaxSellPrice)

	var maxSell float64
	var maxSellSqlNullable sql.NullFloat64
	if err = maxSellRow.Scan(&maxSellSqlNullable); err != nil && err != sql.ErrNoRows {
		err = fmt.Errorf("Error scanning max sell row: %s", err)
		return
	} else if err != nil && err == sql.ErrNoRows {
		// this should never happen but hey if it does we should exit, and not error because that would mean no max
		return
	}

	// if nothing came back (no max) then what are we gonna do
	if !maxSellSqlNullable.Valid {
		return
	} else {
		maxSell = maxSellSqlNullable.Float64
	}

	var minBuyRow *sql.Row
	getminBuyPrice := fmt.Sprintf("SELECT MIN(price) FROM %s WHERE side='%s' FOR UPDATE;", le.pair.String(), buySide.String())
	// errors for queryrow are deferred until scan -- this is important, that's why we don't err != nil here
	minBuyRow = tx.QueryRow(getminBuyPrice)

	var minBuy float64
	var minBuySqlNullable sql.NullFloat64
	if err = minBuyRow.Scan(&minBuySqlNullable); err != nil && err != sql.ErrNoRows {
		err = fmt.Errorf("Error scanning min buy row: %s", err)
		return
	} else if err != nil && err == sql.ErrNoRows {
		// this should never happen but hey if it does we should exit, and not error because that would mean no min
		return
	}

	// if nothing came back (no min) then what are we gonna do
	if !minBuySqlNullable.Valid {
		return
	} else {
		minBuy = minBuySqlNullable.Float64
	}

	// In our prices, if the min buy < max sell, we start to match orders. Otherwise, we can just quit.
	if minBuy > maxSell {
		return
	}

	// TODO: these two decoding / sorting routines may be able to be done concurrently?

	// this will select all sell side, ordered by price descending and time ascending.
	// this means that the sell orders will be sorted by price first, so the best prices will match first,
	// and within the best price the earliest prices will match first.
	var sellRows *sql.Rows
	getSellSideQuery := fmt.Sprintf("SELECT pubkey, price, orderID, amountHave, amountWant, time FROM %s WHERE price>=%f AND side='%s' ORDER BY price DESC, time ASC FOR UPDATE;", le.pair.String(), minBuy, sellSide.String())
	if sellRows, err = tx.Query(getSellSideQuery); err != nil {
		err = fmt.Errorf("Error querying for sell orders for MatchLimitOrders: %s", err)
		return
	}

	// First let's get all the sell orders sorted
	var sellOrders []*match.LimitOrderIDPair
	for sellRows.Next() {
		var pubkeyBytes []byte
		var orderIDBytes []byte
		var timeString string
		sellOrderIDPair := &match.LimitOrderIDPair{
			Order:   new(match.LimitOrder),
			OrderID: new(match.OrderID),
		}
		if err = sellRows.Scan(&pubkeyBytes, &sellOrderIDPair.Price, &orderIDBytes, &sellOrderIDPair.Order.AmountHave, &sellOrderIDPair.Order.AmountWant, &timeString); err != nil {
			err = fmt.Errorf("Error scanning sell rows for MatchLimitOrders: %s", err)
			return
		}

		if sellOrderIDPair.Timestamp, err = time.Parse(sqlTimeFormat, timeString); err != nil {
			err = fmt.Errorf("Error parsing timestamp for MatchLimitOrders: %s", err)
			return
		}

		// we have to do this because ugh they return my byte arrays as hex strings...
		if pubkeyBytes, err = hex.DecodeString(string(pubkeyBytes)); err != nil {
			err = fmt.Errorf("Error decoding hex for sell pubkey for MatchLimitOrders: %s", err)
			return
		}

		// We prepared for this and made a type that knows what's coming with SQL, so we don't
		// have to do the above
		if err = sellOrderIDPair.OrderID.UnmarshalText(orderIDBytes); err != nil {
			err = fmt.Errorf("Error unmarshalling sellorder id for MatchLimitOrders: %s", err)
			return
		}

		sellOrderIDPair.Order.TradingPair = *le.pair
		sellOrderIDPair.Order.Side = match.Sell
		copy(sellOrderIDPair.Order.Pubkey[:], pubkeyBytes)
		sellOrders = append(sellOrders, sellOrderIDPair)
	}
	if err = sellRows.Close(); err != nil {
		err = fmt.Errorf("Error closing sell rows for MatchLimitOrders: %s", err)
		return
	}

	// this will select all buy side, ordered by price ascending and time ascending.
	// this means that the buy orders will be sorted by price first, so the best prices will match first,
	// and within the best price the earliest prices will match first.
	var buyRows *sql.Rows
	getBuySideQuery := fmt.Sprintf("SELECT pubkey, price, orderID, amountHave, amountWant, time FROM %s WHERE price<=%f AND side='%s' ORDER BY price ASC, time ASC FOR UPDATE;", le.pair.String(), maxSell, buySide.String())
	if buyRows, err = tx.Query(getBuySideQuery); err != nil {
		err = fmt.Errorf("Error querying for buy orders for MatchLimitOrders: %s", err)
		return
	}

	// Now let's get all the buy orders sorted
	var buyOrders []*match.LimitOrderIDPair
	for buyRows.Next() {
		var pubkeyBytes []byte
		var orderIDBytes []byte
		var timeString string
		buyOrderIDPair := &match.LimitOrderIDPair{
			Order:   new(match.LimitOrder),
			OrderID: new(match.OrderID),
		}
		if err = buyRows.Scan(&pubkeyBytes, &buyOrderIDPair.Price, &orderIDBytes, &buyOrderIDPair.Order.AmountHave, &buyOrderIDPair.Order.AmountWant, &timeString); err != nil {
			err = fmt.Errorf("Error scanning buy rows for MatchLimitOrders: %s", err)
			return
		}

		if buyOrderIDPair.Timestamp, err = time.Parse(sqlTimeFormat, timeString); err != nil {
			err = fmt.Errorf("Error parsing timestamp for MatchLimitOrders: %s", err)
			return
		}

		// we have to do this because ugh they return my byte arrays as hex strings...
		if pubkeyBytes, err = hex.DecodeString(string(pubkeyBytes)); err != nil {
			err = fmt.Errorf("Error decoding hex for buy pubkey for MatchLimitOrders: %s", err)
			return
		}
		// We prepared for this and made a type that knows what's coming with SQL, so we don't
		// have to do the above
		if err = buyOrderIDPair.OrderID.UnmarshalText(orderIDBytes); err != nil {
			err = fmt.Errorf("Error unmarshalling buyorder id for MatchLimitOrders: %s", err)
			return
		}

		buyOrderIDPair.Order.TradingPair = *le.pair
		buyOrderIDPair.Order.Side = match.Buy
		copy(buyOrderIDPair.Order.Pubkey[:], pubkeyBytes)
		buyOrders = append(buyOrders, buyOrderIDPair)
	}
	if err = buyRows.Close(); err != nil {
		err = fmt.Errorf("Error closing buy rows for MatchLimitOrders: %s", err)
		return
	}

	if orderExecs, settlementExecs, err = match.MatchPrioritizedOrders(buyOrders, sellOrders); err != nil {
		err = fmt.Errorf("Error matching prioritized orders for MatchLimitOrders: %s", err)
		return
	}

	// Update the matching engine with the new state because that's what we do
	for _, orderExec := range orderExecs {
		if orderExec.Filled {
			cancelOrderQuery := fmt.Sprintf("DELETE FROM %s WHERE orderID='%x';", le.pair.String(), orderExec.OrderID)
			if _, err = tx.Exec(cancelOrderQuery); err != nil {
				err = fmt.Errorf("Error deleting filled order for MatchLimitOrders: %s", err)
				return
			}
		} else {
			updateOrderExecQuery := fmt.Sprintf("UPDATE %s SET amountWant='%d', amountHave='%d' WHERE orderID='%x';", le.pair.String(), orderExec.NewAmountWant, orderExec.NewAmountHave, orderExec.OrderID)
			if _, err = tx.Exec(updateOrderExecQuery); err != nil {
				err = fmt.Errorf("Error updating order for order exec for MatchLimitOrders: %s", err)
				return
			}
		}
	}

	return
}

// CreateLimitEngineMap creates a map of pair to limit engine, given a list of pairs.
func CreateLimitEngineMap(pairList []*match.Pair) (limMap map[match.Pair]match.LimitEngine, err error) {

	limMap = make(map[match.Pair]match.LimitEngine)
	var curLimEng match.LimitEngine
	for _, pair := range pairList {
		if curLimEng, err = CreateLimitEngine(pair); err != nil {
			err = fmt.Errorf("Error creating single limit engine while creating limit engine map: %s", err)
			return
		}
		limMap[*pair] = curLimEng
	}

	return
}
