package link

import (
	"go-adv/internal/stat"
	"gorm.io/gorm"
	"math/rand/v2"
)

type Link struct {
	gorm.Model
	Url  string      `json:"url"`
	Hash string      `json:"hash" gorm:"uniqueIndex"`
	Stat []stat.Stat `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func NewLink(url string) *Link {
	link := &Link{
		Url: url,
	}
	link.GenerateHash()
	return link
}

func (link *Link) GenerateHash() {
	link.Hash = RandStringRunes(6)
}

var letterRunes = []rune("abcdefghijklmoprstuvwxyzABCDEFGHIJKLMOPRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.IntN(len(letterRunes))]
	}
	return string(b)
}
