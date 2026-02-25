package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/kroulersama/goProject/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEmployee_Success(t *testing.T) {
	repo, server := setupTestHandler(t)
	defer server.Close()

	// Создаем отдел
	dept := &models.Department{Name: "IT"}
	repo.DB.Create(dept)

	// Создаем сотрудника
	reqBody := models.EmployeeRequest{
		FullName: "John Doe",
		Position: "Developer",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(server.URL+"/departments/1/employees", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "employee created successfully", response["message"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "John Doe", data["full_name"])
	assert.Equal(t, "Developer", data["position"])
}

func TestCreateEmployee_EmptyName(t *testing.T) {
	repo, server := setupTestHandler(t)
	defer server.Close()

	dept := &models.Department{Name: "IT"}
	repo.DB.Create(dept)

	reqBody := models.EmployeeRequest{
		FullName: "",
		Position: "Developer",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(server.URL+"/departments/1/employees", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateEmployee_DepartmentNotFound(t *testing.T) {
	_, server := setupTestHandler(t)
	defer server.Close()

	reqBody := models.EmployeeRequest{
		FullName: "John Doe",
		Position: "Developer",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(server.URL+"/departments/999/employees", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCreateEmployee_HiredAtFuture(t *testing.T) {
	repo, server := setupTestHandler(t)
	defer server.Close()

	dept := &models.Department{Name: "IT"}
	repo.DB.Create(dept)

	futureDate := time.Now().Add(24 * time.Hour)
	reqBody := models.EmployeeRequest{
		FullName: "John Doe",
		Position: "Developer",
		HiredAt:  &futureDate,
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(server.URL+"/departments/1/employees", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
