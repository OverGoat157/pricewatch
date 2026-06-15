package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const wbDetailURL = "https://card.wb.ru/cards/v2/detail"

var (
	wbCatalogRe = regexp.MustCompile(`catalog/(\d+)`)
	digitsRe    = regexp.MustCompile(`^\d+$`)
)

// WB — парсер Wildberries через публичный JSON API карточки товара.
type WB struct {
	client  *http.Client
	baseURL string // вынесен в поле, чтобы подменять в тестах
}

func NewWB() *WB {
	return &WB{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: wbDetailURL,
	}
}

func (w *WB) Name() string { return "wildberries" }

// ExternalID достаёт артикул (nmId) из ссылки вида
// https://www.wildberries.ru/catalog/179978204/detail.aspx или из чистого числа.
func (w *WB) ExternalID(input string) (string, bool) {
	input = strings.TrimSpace(input)
	if m := wbCatalogRe.FindStringSubmatch(input); m != nil {
		return m[1], true
	}
	if digitsRe.MatchString(input) {
		return input, true
	}
	return "", false
}

func (w *WB) Fetch(ctx context.Context, externalID string) (ProductInfo, error) {
	reqURL := fmt.Sprintf("%s?appType=1&curr=rub&dest=-1257786&nm=%s", w.baseURL, url.QueryEscape(externalID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return ProductInfo{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return ProductInfo{}, fmt.Errorf("wb: запрос: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ProductInfo{}, fmt.Errorf("wb: неожиданный статус %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return ProductInfo{}, fmt.Errorf("wb: чтение ответа: %w", err)
	}
	return parseWB(body, externalID)
}

// --- разбор ответа (вынесен отдельно, тестируется на фикстуре) ---

type wbResponse struct {
	Data struct {
		Products []wbProduct `json:"products"`
	} `json:"data"`
}

type wbProduct struct {
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	PriceU        *int64  `json:"priceU"`
	SalePriceU    *int64  `json:"salePriceU"`
	TotalQuantity *int64  `json:"totalQuantity"`
	Sizes         []wbSize `json:"sizes"`
}

type wbSize struct {
	Price *struct {
		Basic   *int64 `json:"basic"`
		Product *int64 `json:"product"`
		Total   *int64 `json:"total"`
	} `json:"price"`
	Stocks []struct {
		Qty int64 `json:"qty"`
	} `json:"stocks"`
}

func parseWB(body []byte, externalID string) (ProductInfo, error) {
	var r wbResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return ProductInfo{}, fmt.Errorf("wb: разбор JSON: %w", err)
	}
	if len(r.Data.Products) == 0 {
		return ProductInfo{}, fmt.Errorf("wb: товар %s не найден", externalID)
	}

	p := r.Data.Products[0]
	price := resolvePrice(p)
	if price <= 0 {
		return ProductInfo{}, fmt.Errorf("wb: не удалось определить цену товара %s", externalID)
	}

	id := p.ID
	if id == 0 {
		id, _ = strconv.ParseInt(externalID, 10, 64)
	}

	return ProductInfo{
		Title:       strings.TrimSpace(p.Name),
		Price:       price,
		IsAvailable: resolveAvailable(p, price),
		ImageURL:    wbImageURL(id),
	}, nil
}

// resolvePrice выбирает первую доступную цену в копейках.
// Формат ответа WB менялся, поэтому проверяем несколько полей по приоритету:
// итоговая цена со скидкой → цена карточки → старые поля salePriceU/priceU → базовая.
func resolvePrice(p wbProduct) int64 {
	if len(p.Sizes) > 0 && p.Sizes[0].Price != nil {
		pr := p.Sizes[0].Price
		for _, v := range []*int64{pr.Total, pr.Product} {
			if v != nil && *v > 0 {
				return *v
			}
		}
	}
	for _, v := range []*int64{p.SalePriceU, p.PriceU} {
		if v != nil && *v > 0 {
			return *v
		}
	}
	if len(p.Sizes) > 0 && p.Sizes[0].Price != nil && p.Sizes[0].Price.Basic != nil {
		return *p.Sizes[0].Price.Basic
	}
	return 0
}

func resolveAvailable(p wbProduct, price int64) bool {
	if p.TotalQuantity != nil {
		return *p.TotalQuantity > 0
	}
	for _, s := range p.Sizes {
		if len(s.Stocks) > 0 {
			return true
		}
	}
	return price > 0
}

// wbImageURL вычисляет ссылку на картинку товара по схеме basket-хостов WB.
// Best-effort: если хост не угадан, картинка просто не загрузится (не критично).
func wbImageURL(nm int64) string {
	if nm <= 0 {
		return ""
	}
	vol := nm / 100000
	part := nm / 1000
	basket := wbBasket(vol)
	return fmt.Sprintf("https://basket-%s.wbbasket.ru/vol%d/part%d/%d/images/big/1.webp", basket, vol, part, nm)
}

func wbBasket(vol int64) string {
	switch {
	case vol <= 143:
		return "01"
	case vol <= 287:
		return "02"
	case vol <= 431:
		return "03"
	case vol <= 719:
		return "04"
	case vol <= 1007:
		return "05"
	case vol <= 1061:
		return "06"
	case vol <= 1115:
		return "07"
	case vol <= 1169:
		return "08"
	case vol <= 1313:
		return "09"
	case vol <= 1601:
		return "10"
	case vol <= 1655:
		return "11"
	case vol <= 1919:
		return "12"
	case vol <= 2045:
		return "13"
	case vol <= 2189:
		return "14"
	case vol <= 2405:
		return "15"
	case vol <= 2621:
		return "16"
	case vol <= 2837:
		return "17"
	case vol <= 3053:
		return "18"
	case vol <= 3269:
		return "19"
	case vol <= 3485:
		return "20"
	case vol <= 3701:
		return "21"
	case vol <= 3917:
		return "22"
	case vol <= 4133:
		return "23"
	case vol <= 4349:
		return "24"
	case vol <= 4565:
		return "25"
	default:
		return "26"
	}
}
