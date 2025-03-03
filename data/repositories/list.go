package repositories

import (
	"beam/config"
	"beam/data/models"
	"errors"
	"time"

	"gorm.io/gorm"
)

type ListRepository interface {
	CheckFavesLine(customerID int, variantID int) (bool, *models.FavesLine, error)
	CheckSavesList(customerID int, variantID int) (bool, *models.SavesList, error)
	CheckLastOrdersList(customerID int, variantID int) (bool, *models.LastOrdersList, error)
	CheckLastOrdersListProd(customerID int, productID int) (bool, *models.LastOrdersList, error)

	CheckCustomListCount(customerID int) (bool, error)
	CreateCustomList(customerID int, name string) (int, error)
	UpdateCustomListTitle(listID int, customerID int, name string) error
	ArchiveCustomList(listID int, customerID int) error

	AddFavesLine(customerID, productID, variantID int, usesDate bool, dateUsed time.Time) error
	AddSavesList(customerID, productID, variantID int, usesDate bool, dateUsed time.Time) error
	AddLastOrdersList(customerID int, orderDate time.Time, orderID string, variants map[int]int) error // Internal?

	CheckLastOrdersListMultiVar(customerID int, variantIDs []int) (map[int]bool, error) // Internal?

	DeleteLastOrdersListVariants(customerID int, variantIDs []int) error // Internal?
	DeleteFavesLine(customerID, variantID int) (time.Time, bool, error)
	DeleteSavesList(customerID, variantID int) (time.Time, bool, error)

	UpdateLastOrdersList(customerID int, orderDate time.Time, orderID string, variants map[int]int) error

	GetFavesListByPage(customerID, page int) ([]*models.FavesLine, bool, bool, error)
	GetSavesListByPage(customerID, page int) ([]*models.SavesList, bool, bool, error)
	GetLastOrdersListByPage(customerID, page int) ([]*models.LastOrdersList, bool, bool, error)
	GetCustomListLineByPage(customerID, page, listID int) ([]*models.CustomListLine, bool, bool, error)

	GetSingleCustomList(customerID, listID int) (*models.CustomList, error)
	AddToCustomList(customerID, listID, variantID, productID int, usesDate bool, dateUsed time.Time) error
	DeleteFromCustomList(listID, customerID, variantID int) (time.Time, bool, error)

	GetCustomListsForCustomer(customerID int) ([]models.CustomList, error)
	CountsForCustomLists(customerID int, listIDs []int) (map[int]int, error)
	HasVariantInLists(customerID, variantID int, listIDs []int) (map[int]bool, error)

	GetFavesListCount(customerID int) (int, error)
	GetSavesListCount(customerID int) (int, error)
	GetLastOrderListCount(customerID int) (int, error)
	GetCustomListCount(customerID, listID int) (int, error)

	SetCustomLastUpdated(customerID, listID int) error
	SetCustomPublicStatus(customerID, listID int, public bool) error
}

type listRepo struct {
	db *gorm.DB
}

func NewListRepository(db *gorm.DB) ListRepository {
	return &listRepo{db: db}
}

func (r *listRepo) CheckFavesLine(customerID int, variantID int) (bool, *models.FavesLine, error) {
	var favesLine models.FavesLine
	if err := r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).First(&favesLine).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil, nil
		}
		return false, nil, err
	}
	return false, &favesLine, nil
}

func (r *listRepo) CheckSavesList(customerID int, variantID int) (bool, *models.SavesList, error) {
	var savesList models.SavesList
	if err := r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).First(&savesList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil, nil
		}
		return false, nil, err
	}
	return false, &savesList, nil
}

func (r *listRepo) CheckLastOrdersList(customerID int, variantID int) (bool, *models.LastOrdersList, error) {
	var lastOrdersList models.LastOrdersList
	if err := r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).
		Order("last_order DESC").
		First(&lastOrdersList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil, nil
		}
		return false, nil, err
	}
	return false, &lastOrdersList, nil
}

func (r *listRepo) CheckLastOrdersListProd(customerID int, productID int) (bool, *models.LastOrdersList, error) {
	var lastOrdersList models.LastOrdersList
	if err := r.db.Where("customer_id = ? AND product_id = ?", customerID, productID).
		Order("last_order DESC").
		First(&lastOrdersList).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil, nil
		}
		return false, nil, err
	}
	return false, &lastOrdersList, nil
}

func (r *listRepo) CheckLastOrdersListMultiVar(customerID int, variantIDs []int) (map[int]bool, error) {
	var existingVariants []int
	result := make(map[int]bool, len(variantIDs))

	if err := r.db.Model(&models.LastOrdersList{}).
		Select("variant_id").
		Where("customer_id = ? AND variant_id IN (?)", customerID, variantIDs).
		Find(&existingVariants).Error; err != nil {
		return nil, err
	}

	for _, id := range variantIDs {
		result[id] = false
	}
	for _, id := range existingVariants {
		result[id] = true
	}

	return result, nil
}

func (r *listRepo) AddFavesLine(customerID, productID, variantID int, usesDate bool, dateUsed time.Time) error {
	var existingFavesLine models.FavesLine
	if err := r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).First(&existingFavesLine).Error; err == nil {
		return nil
	}
	favesLine := models.FavesLine{
		CustomerID: customerID,
		ProductID:  productID,
		VariantID:  variantID,
	}
	if usesDate && dateUsed != (time.Time{}) {
		favesLine.AddDate = dateUsed
	} else {
		favesLine.AddDate = time.Now()
	}
	return r.db.Create(&favesLine).Error
}

func (r *listRepo) AddSavesList(customerID, productID, variantID int, usesDate bool, dateUsed time.Time) error {
	var existingSavesList models.SavesList
	if err := r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).First(&existingSavesList).Error; err == nil {
		if err := r.db.Delete(&existingSavesList).Error; err != nil {
			return err
		}
	}
	savesList := models.SavesList{
		CustomerID: customerID,
		ProductID:  productID,
		VariantID:  variantID,
	}
	if usesDate && dateUsed != (time.Time{}) {
		savesList.AddDate = dateUsed
	} else {
		savesList.AddDate = time.Now()
	}
	return r.db.Create(&savesList).Error
}

func (r *listRepo) AddLastOrdersList(customerID int, orderDate time.Time, orderID string, variants map[int]int) error {
	var lastOrdersList []models.LastOrdersList
	for variantID, productID := range variants {
		lastOrdersList = append(lastOrdersList, models.LastOrdersList{
			CustomerID:  customerID,
			ProductID:   productID,
			VariantID:   variantID,
			LastOrder:   orderDate,
			LastOrderID: orderID,
		})
	}
	return r.db.Create(&lastOrdersList).Error
}

// Time deleted entry was added, if it was already deleted, any error
func (r *listRepo) DeleteFavesLine(customerID, variantID int) (time.Time, bool, error) {
	var line models.FavesLine
	err := r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).
		First(&line).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return time.Time{}, true, nil
	} else if err != nil {
		return time.Time{}, false, r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).
			Delete(&models.FavesLine{}).Error
	}

	return line.AddDate, false, r.db.Delete(&line).Error
}

// Time deleted entry was added, if it was already deleted, any error
func (r *listRepo) DeleteSavesList(customerID, variantID int) (time.Time, bool, error) {
	var line models.SavesList
	err := r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).
		First(&line).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return time.Time{}, true, nil
	} else if err != nil {
		return time.Time{}, false, r.db.Where("customer_id = ? AND variant_id = ?", customerID, variantID).
			Delete(&models.SavesList{}).Error
	}

	return line.AddDate, false, r.db.Delete(&line).Error
}

// Time deleted entry was added, if it was already deleted, any error
func (r *listRepo) DeleteFromCustomList(listID, customerID, variantID int) (time.Time, bool, error) {
	var line models.CustomListLine
	err := r.db.Where("custom_list_id = ? AND customer_id = ? AND variant_id = ?", listID, customerID, variantID).
		First(&line).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return time.Time{}, true, nil
	} else if err != nil {
		return time.Time{}, false, r.db.Where("custom_list_id = ? AND customer_id = ? AND variant_id = ?", listID, customerID, variantID).
			Delete(&models.CustomListLine{}).Error
	}

	return line.AddDate, false, r.db.Delete(&line).Error
}

func (r *listRepo) DeleteLastOrdersListVariants(customerID int, variantIDs []int) error {
	return r.db.Where("customer_id = ? AND variant_id IN (?)", customerID, variantIDs).
		Delete(&models.LastOrdersList{}).Error
}

func (r *listRepo) UpdateLastOrdersList(customerID int, orderDate time.Time, orderID string, variants map[int]int) error {
	variantIDs := make([]int, 0, len(variants))
	for variantID := range variants {
		variantIDs = append(variantIDs, variantID)
	}

	existingMap, err := r.CheckLastOrdersListMultiVar(customerID, variantIDs)
	if err != nil {
		return err
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &listRepo{db: tx}

		toDelete := make([]int, 0)
		for id, exists := range existingMap {
			if exists {
				toDelete = append(toDelete, id)
			}
		}

		if len(toDelete) > 0 {
			if err := txRepo.DeleteLastOrdersListVariants(customerID, toDelete); err != nil {
				return err
			}
		}

		if err := txRepo.AddLastOrdersList(customerID, orderDate, orderID, variants); err != nil {
			return err
		}

		return nil
	})
}

func (r *listRepo) GetFavesListByPage(customerID, page int) ([]*models.FavesLine, bool, bool, error) {
	limit := config.FAVES_LIMIT
	offset := (page - 1)
	if offset < 0 {
		offset = 0
	}
	offset *= limit

	var faves []*models.FavesLine
	if err := r.db.Where("customer_id = ?", customerID).
		Order("add_date DESC").
		Limit(limit + 1).
		Offset(offset).
		Find(&faves).Error; err != nil {
		return nil, false, false, err
	}

	hasPrev := offset > 0
	hasNext := len(faves) > limit
	if hasNext {
		faves = faves[:limit]
	}

	return faves, hasPrev, hasNext, nil
}

func (r *listRepo) GetSavesListByPage(customerID, page int) ([]*models.SavesList, bool, bool, error) {
	limit := config.SAVES_LIMIT
	offset := (page - 1)
	if offset < 0 {
		offset = 0
	}
	offset *= limit

	var saves []*models.SavesList
	if err := r.db.Where("customer_id = ?", customerID).
		Order("add_date DESC").
		Limit(limit + 1).
		Offset(offset).
		Find(&saves).Error; err != nil {
		return nil, false, false, err
	}

	hasPrev := offset > 0
	hasNext := len(saves) > limit
	if hasNext {
		saves = saves[:limit]
	}

	return saves, hasPrev, hasNext, nil
}

func (r *listRepo) GetLastOrdersListByPage(customerID, page int) ([]*models.LastOrdersList, bool, bool, error) {
	limit := config.LAST_ORDERED_LIMIT
	offset := (page - 1)
	if offset < 0 {
		offset = 0
	}
	offset *= limit

	var orders []*models.LastOrdersList
	if err := r.db.Where("customer_id = ?", customerID).
		Order("last_order DESC").
		Limit(limit + 1).
		Offset(offset).
		Find(&orders).Error; err != nil {
		return nil, false, false, err
	}

	hasPrev := offset > 0
	hasNext := len(orders) > limit
	if hasNext {
		orders = orders[:limit]
	}

	return orders, hasPrev, hasNext, nil
}

func (r *listRepo) GetCustomListLineByPage(customerID, page, listID int) ([]*models.CustomListLine, bool, bool, error) {
	limit := config.CUSTOM_LIST_LIMIT
	offset := (page - 1)
	if offset < 0 {
		offset = 0
	}
	offset *= limit

	var lines []*models.CustomListLine
	if err := r.db.Where("customer_id = ?", customerID).
		Order("add_date DESC").
		Limit(limit + 1).
		Offset(offset).
		Find(&lines).Error; err != nil {
		return nil, false, false, err
	}

	hasPrev := offset > 0
	hasNext := len(lines) > limit
	if hasNext {
		lines = lines[:limit]
	}

	return lines, hasPrev, hasNext, nil
}

func (r *listRepo) CheckCustomListCount(customerID int) (bool, error) {
	var count int64
	r.db.Model(&models.CustomList{}).
		Where("customer_id = ? AND archived = false", customerID).
		Count(&count)

	return count < config.MAX_CUSTOM_LISTS, nil
}

func (r *listRepo) CreateCustomList(customerID int, name string) (int, error) {
	ret := models.CustomList{
		CustomerID:  customerID,
		Title:       name,
		Created:     time.Now(),
		LastUpdated: time.Now(),
	}

	if err := r.db.Save(&ret).Error; err != nil {
		return 0, err
	}

	return ret.ID, nil
}

func (r *listRepo) UpdateCustomListTitle(listID int, customerID int, name string) error {
	var customList models.CustomList
	err := r.db.Where("id = ? AND customer_id = ?", listID, customerID).First(&customList).Error
	if err != nil {
		return err
	}

	customList.Title = name
	customList.LastUpdated = time.Now()

	return r.db.Save(&customList).Error
}

func (r *listRepo) ArchiveCustomList(listID int, customerID int) error {
	var customList models.CustomList
	err := r.db.Where("id = ? AND customer_id = ?", listID, customerID).First(&customList).Error
	if err != nil {
		return err
	}

	customList.Archived = true
	customList.ArchivedTime = time.Now()

	return r.db.Save(&customList).Error
}

func (r *listRepo) GetSingleCustomList(customerID, listID int) (*models.CustomList, error) {
	var customList models.CustomList
	err := r.db.Where("id = ? AND customer_id = ? AND archived = false", listID, customerID).First(&customList).Error
	if err != nil {
		return nil, err
	}
	return &customList, nil
}

func (r *listRepo) AddToCustomList(customerID, listID, variantID, productID int, usesDate bool, dateUsed time.Time) error {
	var existingLine models.CustomListLine
	if err := r.db.Where("custom_list_id = ? AND customer_id = ? AND variant_id = ?", listID, customerID, variantID).First(&existingLine).Error; err == nil {
		return nil
	}
	newLine := models.CustomListLine{
		CustomListID: listID,
		CustomerID:   customerID,
		ProductID:    productID,
		VariantID:    variantID,
	}
	if usesDate && dateUsed != (time.Time{}) {
		newLine.AddDate = dateUsed
	} else {
		newLine.AddDate = time.Now()
	}
	return r.db.Create(&newLine).Error
}

func (r *listRepo) GetCustomListsForCustomer(customerID int) ([]models.CustomList, error) {
	var lists []models.CustomList
	r.db.Where("customer_id = ? AND archived = false", customerID).
		Order("last_updated DESC").
		Limit(15).
		Find(&lists)
	return lists, nil
}

func (r *listRepo) CountsForCustomLists(customerID int, listIDs []int) (map[int]int, error) {
	var lineCounts []struct {
		CustomListID int
		LineCount    int
	}

	err := r.db.Model(&models.CustomListLine{}).
		Select("custom_list_id, COUNT(*) AS line_count").
		Where("customer_id = ? AND custom_list_id IN ?", customerID, listIDs).
		Group("custom_list_id").
		Scan(&lineCounts).Error
	if err != nil {
		return nil, err
	}

	countMap := make(map[int]int, len(listIDs))
	for _, lc := range lineCounts {
		countMap[lc.CustomListID] = lc.LineCount
	}

	return countMap, nil
}

func (r *listRepo) HasVariantInLists(customerID, variantID int, listIDs []int) (map[int]bool, error) {
	var results []struct {
		CustomListID int
	}

	err := r.db.Model(&models.CustomListLine{}).
		Select("DISTINCT custom_list_id").
		Where("customer_id = ? AND variant_id = ? AND custom_list_id IN ?", customerID, variantID, listIDs).
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	presenceMap := make(map[int]bool, len(listIDs))
	for _, id := range listIDs {
		presenceMap[id] = false
	}
	for _, res := range results {
		presenceMap[res.CustomListID] = true
	}

	return presenceMap, nil
}

func (r *listRepo) GetFavesListCount(customerID int) (int, error) {
	var count int64
	err := r.db.Model(&models.FavesLine{}).
		Where("customer_id = ?", customerID).
		Distinct("variant_id").
		Count(&count).Error
	return int(count), err
}

func (r *listRepo) GetSavesListCount(customerID int) (int, error) {
	var count int64
	err := r.db.Model(&models.SavesList{}).
		Where("customer_id = ?", customerID).
		Distinct("variant_id").
		Count(&count).Error
	return int(count), err
}

func (r *listRepo) GetLastOrderListCount(customerID int) (int, error) {
	var count int64
	err := r.db.Model(&models.LastOrdersList{}).
		Where("customer_id = ?", customerID).
		Distinct("variant_id").
		Count(&count).Error
	return int(count), err
}

func (r *listRepo) GetCustomListCount(customerID, listID int) (int, error) {
	var count int64
	err := r.db.Model(&models.CustomListLine{}).
		Where("customer_id = ? AND custom_list_id = ?", customerID, listID).
		Distinct("variant_id").
		Count(&count).Error
	return int(count), err
}

func (r *listRepo) SetCustomLastUpdated(customerID, listID int) error {
	return r.db.Model(&models.CustomList{}).
		Where("id = ? AND customer_id = ?", listID, customerID).
		Update("last_updated", time.Now()).Error
}

func (r *listRepo) SetCustomPublicStatus(customerID, listID int, public bool) error {
	return r.db.Model(&models.CustomList{}).
		Where("id = ? AND customer_id = ?", listID, customerID).
		Update("public", public).Error
}
