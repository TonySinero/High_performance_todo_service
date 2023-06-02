package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"newFeatures/models"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/google/uuid"
)

type ElasticSearch struct {
	client *elasticsearch.Client
	index  string
}

func NewTodoElasticSearch(es *elasticsearch.Client, index string) *ElasticSearch {
	return &ElasticSearch{
		client: es,
		index:  index,
	}
}

type todoHits struct {
	Hits struct {
		Hits []struct {
			ID     string             `json:"_id"`
			Source models.TodoElastic `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (e *ElasticSearch) CreateTodo(ctx context.Context, todo *models.TodoElastic) (string, error) {
	// Generate unique ID
	todo.ID = uuid.New().String()

	// Create document in Elasticsearch
	doc, err := json.Marshal(todo)
	if err != nil {
		return "", err
	}
	res, err := e.client.Create(
		e.index,
		todo.ID,
		bytes.NewReader(doc),
		e.client.Create.WithContext(ctx),
	)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.IsError() {
		return "", fmt.Errorf("failed to create document: %s", res.String())
	}

	return todo.ID, nil
}

func (e *ElasticSearch) GetTodoByID(ctx context.Context, id string) (*models.TodoElastic, error) {
	request := esapi.GetRequest{Index: e.index, DocumentID: id}
	response, err := request.Do(ctx, e.client)
	if err != nil {
		return nil, err
	}
	if response.Status() != "200 OK" {
		return nil, errors.New("ElasticSearch: " + response.Status())
	}
	var results Result
	if err := json.NewDecoder(response.Body).Decode(&results); err != nil {
		return nil, err
	}
	results.Source.ID = results.ID
	return &results.Source, nil
}

type Result struct {
	Source models.TodoElastic `json:"_source"`
	ID     string             `json:"_id"`
}

func (e *ElasticSearch) UpdateTodo(ctx context.Context, todo *models.TodoElastic) (string, error) {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(todo); err != nil {
		return "", fmt.Errorf("ElasticSearch update: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      e.index,
		Body:       &buf,
		DocumentID: todo.ID,
		Refresh:    "true",
	}

	resp, err := req.Do(ctx, e.client)
	if err != nil {
		return "", fmt.Errorf("ElasticSearch update: %w", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return "", fmt.Errorf("ElasticSearch update: %w", err)
	}
	return todo.ID, nil
}

func (e *ElasticSearch) GetTodos(ctx context.Context, page, limit int64) ([]models.TodoElastic, error) {
	should := map[string]interface{}{
		"match": map[string]interface{}{
			"_index": e.index,
		},
	}

	query := map[string]interface{}{
		"query": should,
		"size":  limit,
		"from":  (page - 1) * limit,
	}

	hit, err := e.DecodeTodo(ctx, query)
	if err != nil {
		return nil, err
	}

	res := make([]models.TodoElastic, len(hit.Hits.Hits))
	for i, h := range hit.Hits.Hits {
		res[i].ID = h.ID
		res[i].Title = h.Source.Title
		res[i].Completed = h.Source.Completed
	}

	return res, nil
}

func (e *ElasticSearch) SearchTodos(ctx context.Context, query string, page, limit int64) ([]models.TodoElastic, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"title": query,
			},
		},
		"size": limit,
		"from": (page - 1) * limit,
	}

	hit, err := e.DecodeTodo(ctx, searchQuery)
	if err != nil {
		return nil, err
	}

	res := make([]models.TodoElastic, len(hit.Hits.Hits))
	for i, h := range hit.Hits.Hits {
		res[i].ID = h.ID
		res[i].Title = h.Source.Title
		res[i].Completed = h.Source.Completed
	}

	return res, nil
}

func (e *ElasticSearch) DeleteTodoByID(ctx context.Context, id string) error {
	req := esapi.DeleteRequest{
		Index:      e.index,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, e.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ElasticSearch delete: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return err
	}

	if result["result"] != "deleted" {
		return errors.New("ElasticSearch delete: failed to delete document")
	}

	return nil
}

func (e *ElasticSearch) DecodeTodo(ctx context.Context, query map[string]interface{}) (*todoHits, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, err
	}

	req := esapi.SearchRequest{
		Index: []string{e.index},
		Body:  &buf,
	}

	resp, err := req.Do(ctx, e.client)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.IsError() {
		return nil, err
	}
	if resp.Status() != "200 OK" {
		return nil, errors.New("ElasticSearch: " + resp.Status())
	}

	var todo todoHits

	if err := json.NewDecoder(resp.Body).Decode(&todo); err != nil {
		return nil, err
	}
	return &todo, nil
}
