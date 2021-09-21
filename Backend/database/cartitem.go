package database

import (
	"FPproject/Backend/log"
	"FPproject/Backend/models"
	"time"
)

func (d *Database) InsertCI(ci models.CartItem) (string, error) {
	res, err := d.db.Exec("INSERT INTO cart_item(item_id, user_id, qty, remarks, created, updated) VALUES(?,?,?,?,?,?)",
		ci.ID, ci.UserID, ci.Qty, ci.Remarks, time.Now(), time.Now())
	if err != nil {
		log.Warning.Println(err)
		return "", err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		log.Warning.Println(err)
		return "", err
	} else if affected < 1 {
		log.Warning.Println(ErrNoRowsAffected)
		return "", ErrNoRowsAffected
	}
	return ci.ID, nil
}

func (d *Database) DelCI(id string) (string, error) {
	res, err := d.db.Exec("DELETE FROM cart_item WHERE item_id=?", id)
	if err != nil {
		log.Warning.Println(err)
		return "", err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		log.Warning.Println(err)
		return "", err
	} else if affected < 1 {
		log.Warning.Println(ErrNoRowsAffected)
		return "", ErrNoRowsAffected
	}
	return id, nil
}

func (d *Database) UpdateCI(ci models.CartItem) (string, error) {
	res, err := d.db.Exec("UPDATE cart_item SET qty=?, remarks=?, updated=? WHERE item_id=?",
		ci.Qty, ci.Remarks, time.Now(), ci.ID)
	if err != nil {
		log.Warning.Println(err)
		return "", err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		log.Warning.Println(err)
		return "", err
	} else if affected < 1 {
		log.Warning.Println(ErrNoRowsAffected)
		return "", ErrNoRowsAffected
	}
	return ci.ID, nil
}

func (d *Database) GetCI(id string) (models.CartItem, error) {
	var ci models.CartItem
	err := d.db.QueryRow("SELECT * FROM cart_item WHERE item_id=?", id).Scan(&ci.ID,
		&ci.UserID, &ci.Qty, &ci.Remarks, &ci.Created, &ci.Updated)
	if err != nil {
		log.Warning.Println(err)
		return ci, err
	}
	return ci, nil
}