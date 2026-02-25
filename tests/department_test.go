package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/kroulersama/goProject/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDepartment_Success(t *testing.T) {
	_, server := setupTestHandler(t)
	defer server.Close()

	// Подготовка запроса
	reqBody := models.DepartmentRequest{
		Name: "IT Department",
	}
	body, _ := json.Marshal(reqBody)

	// Отправка запроса
	resp, err := http.Post(server.URL+"/departments", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Проверка ответа
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "department created successfully", response["message"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "IT Department", data["name"])
	assert.NotNil(t, data["id"])
}

func TestCreateDepartment_EmptyName(t *testing.T) {
	_, server := setupTestHandler(t)
	defer server.Close()

	reqBody := models.DepartmentRequest{
		Name: "",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(server.URL+"/departments", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var response map[string]string
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, models.ErrNameEmpty.Error(), response["message"])
}

func TestCreateDepartment_DuplicateName(t *testing.T) {
	repo, server := setupTestHandler(t)
	defer server.Close()

	// Создаем первый отдел
	dept1 := &models.Department{Name: "IT"}
	repo.DB.Create(dept1)

	// Пытаемся создать второй с таким же именем
	reqBody := models.DepartmentRequest{
		Name: "IT",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(server.URL+"/departments", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestGetDepartment_Success(t *testing.T) {
	repo, server := setupTestHandler(t)
	defer server.Close()

	// Создаем тестовый отдел
	dept := &models.Department{Name: "Test Dept"}
	repo.DB.Create(dept)

	// Получаем отдел
	resp, err := http.Get(server.URL + "/departments/1")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response models.DepartmentResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Test Dept", response.Name)
}

func TestGetDepartment_NotFound(t *testing.T) {
	_, server := setupTestHandler(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/departments/999")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetDepartment_WithDepth(t *testing.T) {
	repo, server := setupTestHandler(t)
	defer server.Close()

	// Создаем иерархию
	parent := &models.Department{Name: "Parent"}
	repo.DB.Create(parent)

	child := &models.Department{Name: "Child", ParentId: &parent.Id}
	repo.DB.Create(child)

	grandchild := &models.Department{Name: "Grandchild", ParentId: &child.Id}
	repo.DB.Create(grandchild)

	// Запрос с depth=2
	resp, err := http.Get(server.URL + "/departments/1?depth=2")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response models.DepartmentResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Parent", response.Name)
	assert.Len(t, response.Children, 1)
	assert.Equal(t, "Child", response.Children[0].Name)
	// На depth=2 внуков не должно быть
	assert.Empty(t, response.Children[0].Children)
}
