package models

import "time"

type User struct {
	ID        uint   `gorm:"primary_key"`
	Username  string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	CreatedAt time.Time
}

type Document struct {
	ID               uint   `gorm:"primary_key"`
	UserID           uint   `gorm:"not null;index:idx_user_file,priority:1"`
	Filename         string `gorm:"not null;index:idx_user_file,priority:2,unique"`
	Content          string `gorm:"type:text;not null"`
	OriginalPath     string `gorm:"not null"`
	ProcessedContent string `gorm:"type:text;not null"`
	CreatedAt        time.Time

	Collections []*Collection `gorm:"many2many:collection_documents;constraint:OnDelete:CASCADE;"`
}

type Collection struct {
	ID        uint   `gorm:"primary_key"`
	UserID    uint   `gorm:"not null;index"`
	Name      string `gorm:"not null"`
	CreatedAt time.Time

	IDFRecords []CollectionIDF `gorm:"constraint:OnDelete:CASCADE;"`
	Documents  []*Document     `gorm:"many2many:collection_documents;"`
}

type CollectionIDF struct {
	ID           uint    `gorm:"primary_key"`
	CollectionID uint    `gorm:"not null;index"`
	Word         string  `gorm:"not null;index"`
	IDFValue     float64 `gorm:"not null"`
}
