package main

import (
	"context"
	"database/sql"
	"time"

	"github.com/redis/go-redis/v9"
)

func (m *ShortnerModel) Close() {
	m.cancel()
	m.redis.Close()
	m.db.Close()
}

type ShortnerModel struct {
	db     *sql.DB
	redis  *redis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

func Init(db *sql.DB, redis *redis.Client) *ShortnerModel {
	ctx, cancel := context.WithCancel(context.Background())
	return &ShortnerModel{
		db:     db,
		redis:  redis,
		ctx:    ctx,
		cancel: cancel,
	}
}

// add books in db
func (m *ShortnerModel) CreateShortner(long_url string, user_id int) (string, error) {
	var hash_value string
	err := m.db.QueryRow("SELECT `hash` FROM `url_shortner` WHERE  `long_url` = ? AND active = 1", long_url).Scan(&hash_value)
	if err == nil {
		return hash_value, err
	} else if err != sql.ErrNoRows {
		return hash_value, err
	}

	result, err := m.db.Exec("INSERT INTO `url_shortner` (`long_url`,`user_id`) VALUES (?,?)", &long_url, &user_id)
	if err != nil {
		return "", err
	}

	url_id, err := result.LastInsertId()
	if err != nil {
		return "", err
	}

	hash_value = m.base62Encode(url_id)
	_, err = m.db.Exec("UPDATE `url_shortner` SET `hash` = ? ,`active` = 1 WHERE id = ?", hash_value, url_id)
	return hash_value, err
}

// check isbn already exist or not()
func (m *ShortnerModel) GetShortner(hash_value string) (string, error) {
	var long_url string
	err := m.db.QueryRow("SELECT `long_url` FROM `url_shortner` WHERE  `hash` = ? AND active = 1", hash_value).Scan(&long_url)
	if err != nil {
		return "", err
	}

	return long_url, err
}

func (m *ShortnerModel) IncrementHit(hash_value string) error {
	_, err := m.db.Exec("UPDATE `url_shortner` SET `traffic` = `traffic`+1 ,`active` = 1 WHERE `hash` = ?", hash_value)
	return err
}

func (m *ShortnerModel) base62Encode(id int64) string {
	base62Digits := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	url := ""
	i := id
	for i > 0 {
		remanider := i % 62
		url = string(base62Digits[remanider]) + url
		i = i / 62
	}

	return url
}

func (m *ShortnerModel) RedisSet(key, value string) error {
	err := m.redis.Set(m.ctx, key, value, 5*time.Minute).Err()
	m.Close()
	return err
}

func (m *ShortnerModel) RedisGet(hash_value string) (string, error) {
	val, err := m.redis.Get(m.ctx, hash_value).Result()
	return val, err
}
