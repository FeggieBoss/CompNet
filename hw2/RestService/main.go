package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

type item struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var items = []item{
	{ID: "0", Name: "rice", Description: "160g | 544kcal | 15protein"},
	{ID: "1", Name: "buckwheat", Description: "160g | 580kcal | 20protein"},
	{ID: "2", Name: "pasta", Description: "150g | 538kcal | 21protein"},
	{ID: "3", Name: "6fried eggs", Description: "270g | 455kcal | 32protein"},
	{ID: "4", Name: "maasdam cheese", Description: "55g | 185kcal | 14protein"},
}

var itemsImgNames = map[string]string{}

// Получить список всех продуктов
func getItems(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, items)
}

// Добавить новый продукт. При этом его id должен сгенерироваться автоматически
func postItem(c *gin.Context) {
	var newItem item

	if err := c.BindJSON(&newItem); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "incorrect request body format"})
		return
	}

	items = append(items, newItem)

	// генерация id
	if len(items) == 1 {
		items[len(items)-1].ID = "0"
	} else {
		v, _ := strconv.Atoi(items[len(items)-2].ID)
		items[len(items)-1].ID = fmt.Sprint(v + 1)
	}

	c.IndentedJSON(http.StatusCreated, items)
}

// Получить продукт по его id
func getItemByID(c *gin.Context) {
	id := c.Param("id")

	for _, a := range items {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "item not found"})
}

// Обновить существующий продукт (обновленные данные продукта передаются в теле запроса)
func updateItemByID(c *gin.Context) {
	id := c.Param("id")

	for i := 0; i < len(items); i++ {
		if items[i].ID == id {
			var updItem item

			if err := c.BindJSON(&updItem); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"message": "incorrect request body format"})
				return
			}

			if newName := updItem.Name; len(newName) > 0 {
				items[i].Name = newName
			}
			if newDescription := updItem.Description; len(newDescription) > 0 {
				items[i].Description = newDescription
			}
			c.IndentedJSON(http.StatusOK, items)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "item not found"})
}

// Удалить продукт по его id
func deleteItemByID(c *gin.Context) {
	id := c.Param("id")

	for i := 0; i < len(items); i++ {
		if items[i].ID == id {
			items = append(items[:i], items[i+1:]...)
			c.IndentedJSON(http.StatusOK, items)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "item not found"})
}

// Получить картинку продукта
func getImgById(c *gin.Context) {
	id := c.Param("id")

	v, ok := itemsImgNames[id]
	if !ok {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "image not found"})
	}

	c.FileAttachment(filepath.Base(v), v)
}

// Добавить/обновить картинку продукту
func postImg(c *gin.Context) {
	id := c.PostForm("id")

	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "get form err: %s", err.Error())
		return
	}

	filename := filepath.Base(file.Filename)
	if err := c.SaveUploadedFile(file, filename); err != nil {
		c.String(http.StatusBadRequest, "upload file err: %s", err.Error())
		return
	}

	itemsImgNames[id] = file.Filename
	c.String(http.StatusOK, "File %s uploaded successfully with id=%s", file.Filename, id)
}

func deleteImgById(c *gin.Context) {
	id := c.Param("id")

	_, ok := itemsImgNames[id]
	if !ok {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "image not found"})
	}

	delete(itemsImgNames, id)
	c.String(http.StatusOK, "Delete successfully")
}

func main() {
	router := gin.Default()
	router.GET("/items", getItems)
	router.POST("/items", postItem)
	router.GET("/items/:id", getItemByID)
	router.PUT("/items/:id", updateItemByID)
	router.DELETE("/items/:id", deleteItemByID)

	router.GET("/itemsImg/:id", getImgById)
	router.POST("/itemsImg", postImg)
	router.DELETE("/itemsImg/:id", deleteImgById)

	router.Run("localhost:8080")
}
