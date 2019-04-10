package ocxsql

import (
	"database/sql"
	"fmt"

	"github.com/mit-dci/lit/lncore"
)

// GetPeerAddrs gets the peer lnaddresses from the database
func (db *DB) GetPeerAddrs() (lnAddresses []lncore.LnAddr, err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		return
	}

	// Add commit to defer stack
	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while getting peer addrs: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// use peer schema
	if _, err = tx.Exec("USE " + db.peerSchema + ";"); err != nil {
		return
	}

	var rows *sql.Rows
	getPeerAddrsQuery := fmt.Sprintf("SELECT lnaddr FROM %s;", db.peerSchema)
	if rows, err = tx.Query(getPeerAddrsQuery); err != nil {
		return
	}

	// Add row close to defer stack
	defer func() {
		if err = rows.Close(); err != nil {
			err = fmt.Errorf("Error closing peer rows: \n%s", err)
		}
	}()

	for rows.Next() {
		var lnAddrBytes []byte
		if err = rows.Scan(&lnAddrBytes); err != nil {
			err = fmt.Errorf("Error scanning for ln addrs: \n%s", err)
			return
		}
		// This does weird stuff when putting the bytes in but creating a string works

		var lnAddr lncore.LnAddr
		if lnAddr, err = lncore.ParseLnAddr(string(lnAddrBytes)); err != nil {
			return
		}

		lnAddresses = append(lnAddresses, lnAddr)
	}

	return
}

// GetPeerInfo gets the peer info for a specific lnaddr
func (db *DB) GetPeerInfo(addr lncore.LnAddr) (peerInfo lncore.PeerInfo, err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while getting peer info for a single peer: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// use peer schema
	if _, err = tx.Exec("USE " + db.peerSchema + ";"); err != nil {
		return
	}

	var rows *sql.Rows
	getPeerInfoQuery := fmt.Sprintf("SELECT name, netaddr, peerIdx FROM %s WHERE lnaddr='%s';", db.peerSchema, addr.ToString())
	if rows, err = tx.Query(getPeerInfoQuery); err != nil {
		return
	}

	// Add row close to defer stack
	defer func() {
		if err = rows.Close(); err != nil {
			err = fmt.Errorf("Error closing peer rows: \n%s", err)
		}
	}()

	// init the data
	var nickname string
	var netAddr string
	var peerIdx uint32

	// only get one, that's why we use if instead of for
	if rows.Next() {

		if err = rows.Scan(&nickname, &netAddr, &peerIdx); err != nil {
			err = fmt.Errorf("Error scanning for peer info: \n%s", err)
			return
		}

	} else {
		err = fmt.Errorf("No peers found for that address")
		return
	}

	// actually set the data to return
	peerInfo = lncore.PeerInfo{
		LnAddr:   &addr,
		Nickname: &nickname,
		NetAddr:  &netAddr,
		PeerIdx:  peerIdx,
	}

	return
}

// GetPeerInfos gets a map of lnaddr's to peer info
func (db *DB) GetPeerInfos() (peerInfos map[lncore.LnAddr]lncore.PeerInfo, err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while getting all peer infos: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// use peer schema
	if _, err = tx.Exec("USE " + db.peerSchema + ";"); err != nil {
		return
	}

	// TODO: Finish this function
	var rows *sql.Rows
	getPeerInfosQuery := fmt.Sprintf("SELECT lnaddr, name, netaddr, peerIdx FROM %s;", db.peerSchema)
	if rows, err = tx.Query(getPeerInfosQuery); err != nil {
		return
	}

	// Add row close to defer stack
	defer func() {
		if err = rows.Close(); err != nil {
			err = fmt.Errorf("Error closing peer rows: \n%s", err)
		}
	}()

	// make the map
	peerInfos = make(map[lncore.LnAddr]lncore.PeerInfo)

	// only get one, that's why we use if instead of for
	for rows.Next() {
		// init the data
		var lnAddrBytes []byte
		var nickname string
		var netAddr string
		var peerIdx uint32

		if err = rows.Scan(&lnAddrBytes, &nickname, &netAddr, &peerIdx); err != nil {
			err = fmt.Errorf("Error scanning for peer info: \n%s", err)
			return
		}

		var lnAddr lncore.LnAddr
		if lnAddr, err = lncore.ParseLnAddr(string(lnAddrBytes)); err != nil {
			return
		}

		// actually set the data to return
		peerInfos[lnAddr] = lncore.PeerInfo{
			LnAddr:   &lnAddr,
			Nickname: &nickname,
			NetAddr:  &netAddr,
			PeerIdx:  peerIdx,
		}

	}

	return
}

// AddPeer adds a peer (address, peerinfo) to the database
func (db *DB) AddPeer(addr lncore.LnAddr, pi lncore.PeerInfo) (err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while adding peer: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// use peer schema
	if _, err = tx.Exec("USE " + db.peerSchema + ";"); err != nil {
		return
	}

	addPeerQuery := fmt.Sprintf("INSERT INTO %s VALUES ('%s', '%s', '%s', %d);", db.peerSchema, pi.LnAddr.ToString(), *pi.NetAddr, *pi.Nickname, pi.PeerIdx)
	if _, err = tx.Exec(addPeerQuery); err != nil {
		return
	}

	return
}

// UpdatePeer updates a peer with addr with new peerinfo
func (db *DB) UpdatePeer(addr lncore.LnAddr, pi *lncore.PeerInfo) (err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while updating peer: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// use peer schema
	if _, err = tx.Exec("USE " + db.peerSchema + ";"); err != nil {
		return
	}

	// this should not be a bug, if you tell the database to update a lnaddr with a new lnaddr it will do it for you, right?
	updatePeerQuery := fmt.Sprintf("UPDATE %s SET lnaddr='%s', netaddr='%s', name='%s', peerIdx=%d WHERE lnaddr='%s';", db.peerSchema, pi.LnAddr.ToString(), *pi.NetAddr, *pi.Nickname, pi.PeerIdx, addr.ToString())
	if _, err = tx.Exec(updatePeerQuery); err != nil {
		return
	}

	return
}

// DeletePeer deletes a peer from the database
func (db *DB) DeletePeer(addr lncore.LnAddr) (err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while deleting peer: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// use peer schema
	if _, err = tx.Exec("USE " + db.peerSchema + ";"); err != nil {
		return
	}

	// delete the peer from the peer table
	deletePeerQuery := fmt.Sprintf("DELETE FROM %s WHERE lnaddr='%s';", db.peerSchema, addr.ToString())
	if _, err = tx.Exec(deletePeerQuery); err != nil {
		return
	}

	return
}

// GetUniquePeerIdx gets a unique peer index
func (db *DB) GetUniquePeerIdx() (peerIdx uint32, err error) {

	var tx *sql.Tx
	if tx, err = db.DBHandler.Begin(); err != nil {
		return
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			err = fmt.Errorf("Error while getting unique peer index: \n%s", err)
			return
		}
		err = tx.Commit()
	}()

	// use peer schema
	if _, err = tx.Exec("USE " + db.peerSchema + ";"); err != nil {
		return
	}

	var rows *sql.Rows
	getMaxPeerIdxQuery := fmt.Sprintf("SELECT MAX(peerIdx) FROM %s;", db.peerSchema)
	if rows, err = tx.Query(getMaxPeerIdxQuery); err != nil {
		return
	}

	// Add row close to defer stack
	defer func() {
		if err = rows.Close(); err != nil {
			err = fmt.Errorf("Error closing peer rows: \n%s", err)
		}
	}()

	if rows.Next() {
		if err = rows.Scan(&peerIdx); err != nil {
			err = fmt.Errorf("Error scanning for max peer idx: \n%s", err)
			return
		}
	}

	return
}
