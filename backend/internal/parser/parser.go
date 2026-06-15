// Package parser извлекает текущие данные товара из внешних источников.
package parser

import "context"

// ProductInfo — результат парсинга карточки товара.
type ProductInfo struct {
	Title       string
	Price       int64 // копейки
	IsAvailable bool
	ImageURL    string
}

// Parser — источник цен. Реализация подменяется без изменения остального кода
// (сейчас только Wildberries, но можно добавить другие маркетплейсы).
type Parser interface {
	// Name — имя источника, попадает в products.source.
	Name() string
	// ExternalID извлекает идентификатор товара из ссылки или строки с артикулом.
	// ok=false, если строку не удалось распознать.
	ExternalID(input string) (id string, ok bool)
	// Fetch получает актуальные данные товара по его идентификатору.
	Fetch(ctx context.Context, externalID string) (ProductInfo, error)
}
