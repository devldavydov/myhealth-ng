package httpapi

import (
	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/port"
)

type Options struct {
	CertificateRequired bool
}

func NewRouter(foodCases port.FoodUseCases, weightCases port.WeightUseCases, bundleCases port.BundleUseCases, settingsCases port.SettingsUseCases, journalCases port.JournalUseCases, activeCaloriesCases port.ActiveCaloriesUseCases, options Options) *gin.Engine {
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(gin.Logger(), recovery(), identity(options.CertificateRequired))

	food := newFoodHandler(foodCases)
	weight := newWeightHandler(weightCases)
	bundle := newBundleHandler(bundleCases)
	settings := newSettingsHandler(settingsCases)
	journal := newJournalHandler(journalCases)
	activeCalories := newActiveCaloriesHandler(activeCaloriesCases)
	user := newUserHandler()

	router.GET("/api/me", user.getCurrent)
	router.GET("/api/settings", settings.get)
	router.PUT("/api/settings", settings.save)
	router.GET("/api/journal", journal.get)
	router.POST("/api/journal", journal.save)
	router.DELETE("/api/journal/:dt/:meal/:foodKey", journal.deleteItem)
	router.DELETE("/api/journal/:dt/:meal", journal.clearMeal)
	router.GET("/api/active-calories", activeCalories.get)
	router.POST("/api/active-calories", activeCalories.save)
	router.DELETE("/api/active-calories/:dt", activeCalories.delete)
	router.GET("/api/food", food.list)
	router.GET("/api/food/:key", food.get)
	router.POST("/api/food", food.create)
	router.PUT("/api/food/:key", food.update)
	router.DELETE("/api/food/:key", food.delete)
	router.GET("/api/weight", weight.list)
	router.POST("/api/weight", weight.save)
	router.DELETE("/api/weight/:dt", weight.delete)
	router.GET("/api/bundle", bundle.list)
	router.GET("/api/bundle/:key", bundle.get)
	router.POST("/api/bundle", bundle.create)
	router.PUT("/api/bundle/:key", bundle.update)
	router.DELETE("/api/bundle/:key", bundle.delete)
	router.NoRoute(notFound)

	return router
}
