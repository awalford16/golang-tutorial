package data

import (
	"context"
	"fmt"
	"time"

	protos "tutorials/currency/protos/currency"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
)

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"gt=0"`
	CreatedOn   string  `json:"-"`
	UpdatedOn   string  `json:"-"`
}

func (p *Product) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type Products []*Product

type ProductsDB struct {
	currency protos.CurrencyClient
	log      hclog.Logger
}

func NewProductsDB(c protos.CurrencyClient, l hclog.Logger) *ProductsDB {
	return &ProductsDB{c, l}
}

func (p *ProductsDB) GetProducts(currency string) (Products, error) {
	if currency == "" {
		return productList, nil
	}

	// Update product list currencies
	rate, err := p.getRate(currency)
	if err != nil {
		p.log.Error("Unable to get rate", "currency", currency, "error", err)
		return nil, err
	}

	pr := Products{}
	for _, p := range productList {
		np := *p
		np.Price = rate * np.Price
		pr = append(pr, &np)
	}

	return pr, nil
}

func (p *ProductsDB) GetProductByID(id int, currency string) (*Product, error) {
	i := findIndexByProductID(id)
	if i == -1 {
		return nil, ErrProductNotFound
	}

	if currency == "" {
		return productList[i], nil
	}

	rate, err := p.getRate(currency)
	if err != nil {
		p.log.Error("Unable to get rate", "currency", currency, "error", err)
		return nil, err
	}

	// Take copy of product list item
	pc := *productList[i]
	pc.Price = pc.Price * rate

	return &pc, nil
}

func (pd *ProductsDB) AddProduct(p *Product) {
	p.ID = getNextId()

	productList = append(productList, p)
}

func (pd *ProductsDB) UpdateProduct(id int, p *Product) error {
	i := findIndexByProductID(id)
	if i == -1 {
		return ErrProductNotFound
	}

	p.ID = i
	productList[i] = p

	return nil
}

var ErrProductNotFound = fmt.Errorf("Product not found")

func findIndexByProductID(id int) int {
	for i, p := range productList {
		if p.ID == id {
			return i
		}
	}

	return -1
}

func (p *ProductsDB) getRate(destination string) (float64, error) {
	// Get currency conversion
	rr := &protos.RateRequest{
		Base:        protos.Currencies(protos.Currencies_value["EUR"]),
		Destination: protos.Currencies(protos.Currencies_value[destination]),
	}

	resp, err := p.currency.GetRate(context.Background(), rr)
	return resp.Rate, err
}

func getNextId() int {
	lp := productList[len(productList)-1]
	return lp.ID + 1
}

var productList = []*Product{
	&Product{
		ID:          1,
		Name:        "Cappucino",
		Description: "Milky Coffee",
		Price:       3.00,
		CreatedOn:   time.Now().UTC().String(),
		UpdatedOn:   time.Now().UTC().String(),
	},
	&Product{
		ID:          2,
		Name:        "Latte",
		Description: "Another Milky Coffee",
		Price:       2.45,
		CreatedOn:   time.Now().UTC().String(),
		UpdatedOn:   time.Now().UTC().String(),
	},
}
