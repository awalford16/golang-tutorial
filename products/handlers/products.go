package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/awalford16/golang-tutorial/products/data"

	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
)

type Products struct {
	l         hclog.Logger
	productDB *data.ProductsDB
}

func NewProducts(l hclog.Logger, pdb *data.ProductsDB) *Products {
	return &Products{l, pdb}
}

func (p *Products) GetProducts(rw http.ResponseWriter, r *http.Request) {
	cur := r.URL.Query().Get("currency")

	// Get currency from query strings
	lp, err := p.productDB.GetProducts(cur)
	if err != nil {
		http.Error(rw, "Unable to get products", http.StatusInternalServerError)
		return
	}

	err = data.ToJSON(lp, rw)
	if err != nil {
		p.l.Error("Unable to marshal JSON", "error", err)
	}
}

func (p *Products) GetProduct(rw http.ResponseWriter, r *http.Request) {
	id := getProductID(r)
	cur := r.URL.Query().Get("currency")

	p.l.Debug("Get Product with ID", id)

	prod, err := p.productDB.GetProductByID(id, cur)
	switch err {
	case nil:
	case data.ErrProductNotFound:
		rw.WriteHeader(http.StatusNotFound)
		p.l.Error("Product not found")
		return
	default:
		rw.WriteHeader(http.StatusInternalServerError)
		p.l.Error("Error fetching product")
		return
	}

	err = data.ToJSON(prod, rw)
	if err != nil {
		p.l.Error("Error serializing product", err)
	}
}

func (p *Products) AddProduct(rw http.ResponseWriter, r *http.Request) {
	p.l.Debug("Handle Post Product")

	prod := r.Context().Value(KeyProduct{}).(data.Product)

	p.l.Debug("Prod: %#v", &prod)
	p.productDB.AddProduct(&prod)
}

func (p *Products) UpdateProduct(rw http.ResponseWriter, r *http.Request) {
	id := getProductID(r)
	p.l.Debug("Handle PUT Product", id)

	// Cast request product to Product obj
	prod := r.Context().Value(KeyProduct{}).(data.Product)

	err := p.productDB.UpdateProduct(id, &prod)
	if err == data.ErrProductNotFound {
		http.Error(rw, "Product not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(rw, "Failed to update", http.StatusInternalServerError)
		return
	}
}

func getProductID(r *http.Request) int {
	// Get the path variables from request URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		return -1
	}

	return id
}

type KeyProduct struct{}

func (p Products) MiddlewareProductValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		prod := data.Product{}

		err := data.FromJSON(prod, r.Body)
		if err != nil {
			http.Error(rw, "Unable to unmarshal json", http.StatusBadRequest)
			return
		}

		// Add product to request context
		ctx := context.WithValue(r.Context(), KeyProduct{}, prod)
		req := r.WithContext(ctx)

		next.ServeHTTP(rw, req)
	})
}
