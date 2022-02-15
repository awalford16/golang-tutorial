package data

import (
	"context"
	"fmt"
	"time"

	protos "github.com/awalford16/golang-tutorial/currency/protos/currency"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	rates    map[string]float64
	client   protos.Currency_SubscribeRatesClient
}

func NewProductsDB(c protos.CurrencyClient, l hclog.Logger) *ProductsDB {
	pb := &ProductsDB{c, l, make(map[string]float64), nil}

	go pb.handleUpdates()

	return pb
}

func (p *ProductsDB) handleUpdates() {
	sub, err := p.currency.SubscribeRates(context.Background())
	if err != nil {
		p.log.Error("Unable to subscribe for rates", "error", err)
	}

	// Set the client subscription
	p.client = sub

	for {
		rr, err := sub.Recv()

		// Proto defines that the response should be one of message or error
		if grpcErr := rr.GetError(); grpcErr != nil {
			p.log.Error("Error subscribing for rates", "error", grpcErr.Message)
			continue
		}

		if err != nil {
			p.log.Error("Error receiving message", "error", err)
			return
		}

		resp := rr.GetRateResponse()
		p.log.Info("Recieved updated rates from server", "destination", resp.GetDestination().String())
		p.rates[resp.Destination.String()] = resp.Rate
	}
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
	// If rate is cached, return it
	if r, ok := p.rates[destination]; ok {
		return r, nil
	}

	// Get currency conversion
	rr := &protos.RateRequest{
		Base:        protos.Currencies(protos.Currencies_value["EUR"]),
		Destination: protos.Currencies(protos.Currencies_value[destination]),
	}

	// Get initial rate
	resp, err := p.currency.GetRate(context.Background(), rr)
	if err != nil {
		if s, ok := status.FromError(err); ok {
			md := s.Details()[0].(*protos.RateRequest)

			// Handle individual errors from server
			if s.Code() == codes.InvalidArgument {
				return -1, fmt.Errorf("Unable to get rate from currency server, destination and base currencies cannot be the same")
			}
			return -1, fmt.Errorf("Unable to get rate from currency server, base: %s, dest: %s", md.Base.String(), md.Destination.String())
		}

		return -1, err
	}
	p.rates[destination] = resp.Rate // Update cache

	// Subscribe for updates
	p.client.Send(rr)

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
