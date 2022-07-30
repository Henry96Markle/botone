package main

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type (
	Configuration struct {
		BotToken         string `json:"bot_token"`
		OwnerTelegramID  int64  `json:"owner_telegram_id"`
		ConnectionString string `json:"connection_string"`
		LogChannelID     int64  `json:"log_channel_id"`
		LoggingToChannel bool   `json:"logging_to_channel"`
	}

	Record struct {
		ChatID int64     `bson:"chat_id" json:"chat_id"`
		Notes  []string  `bson:"notes" json:"notes"`
		Date   time.Time `bson:"date" json:"date"`
	}

	User struct {
		ID          primitive.ObjectID    `bson:"_id" json:"_id"`
		Names       []string              `bson:"names" json:"names"`
		Usernames   []string              `bson:"usernames" json:"usernames"`
		TelegramID  int64                 `bson:"tg_id" json:"tg_id"`
		AliasIDs    []int64               `bson:"alias_ids" json:"alias_ids"`
		Permission  int                   `bson:"permission_level" json:"permission_level"`
		Description string                `bson:"description" json:"description"`
		Records     map[string]([]Record) `bson:"records" json:"records"`
	}

	// User structure is a wrapper for the MongoDB document.
	Database struct {
		client     *mongo.Client
		database   *mongo.Database
		collection string
	}
)
