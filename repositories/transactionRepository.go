package repositories

import (
	"fmt"
	"math"
	"strings"

	"jwt-authen/database"
	"jwt-authen/dtos"
	"jwt-authen/models"
)

func TransactionPagination(pagination *dtos.Pagination, id uint32) (*dtos.Pagination, int, error) {
	var transactions []models.Transaction

	totalPages, fromRow, toRow := 0, 0, 0

	var totalRows int64 = 0

	offset := (pagination.Page - 1) * pagination.Limit

	find := database.Db.Limit(pagination.Limit).Offset(offset).Order(pagination.Sort)

	searchs := pagination.Searchs

	if searchs != nil {
		for _, value := range searchs {
			column := value.Column
			action := value.Action
			query := value.Query

			switch action {
			case "equals":
				whereQuery := fmt.Sprintf("%s = ?", column)
				find = find.Where(whereQuery, query)
				break
			case "contains":
				whereQuery := fmt.Sprintf("%s LIKE ?", column)
				find = find.Where(whereQuery, "%"+query+"%")
				break
			case "in":
				whereQuery := fmt.Sprintf("%s IN (?)", column)
				queryArray := strings.Split(query, ",")
				find = find.Where(whereQuery, queryArray)
				break
			}
		}
	}

	if err := find.Find(&transactions).Error; err != nil {
		return nil, totalPages, err
	}

	pagination.Rows = transactions

	if err := database.Db.Model(&models.Transaction{}).Count(&totalRows).Error; err != nil {
		return nil, totalPages, err
	}

	pagination.TotalRows = totalRows

	totalPages = int(math.Ceil(float64(totalRows) / float64(pagination.Limit)))

	if pagination.Page == 1 {
		fromRow = 1
		toRow = pagination.Limit
	} else {
		if pagination.Page <= totalPages {
			fromRow = pagination.Page*pagination.Limit + 1
			toRow = (pagination.Page + 1) * pagination.Limit
		}
	}

	if int64(toRow) > totalRows {
		toRow = int(totalRows)
	}

	pagination.FromRow = fromRow
	pagination.ToRow = toRow

	return pagination, totalPages, nil
}
