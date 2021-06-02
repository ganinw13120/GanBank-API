package main

import (
	Route "GANBANKING_API/src/route"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {

	godotenv.Load()
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(200, "Success")
	})

	mobileGroup := e.Group("/mobile")
	cmsGroup := e.Group("cms")

	mobileCustomerGroup := mobileGroup.Group("/customer")
	mobileCustomerGroup.POST("/create", Route.CreateCustomer)
	mobileCustomerGroup.PUT("/update", Route.UpdateCustomer)

	cmsAccountGroup := cmsGroup.Group("/account")
	cmsAccountGroup.POST("/create", Route.CreateAccount)
	cmsAccountGroup.GET("/info", Route.GetAccount)
	cmsAccountGroup.POST("/prepare", Route.PrepareAccount)

	mobileAccountGroup := mobileGroup.Group("/account")
	mobileAccountGroup.POST("/info", Route.GetAccountByID)

	mobileBankGroup := mobileGroup.Group("/bank")
	mobileBankGroup.GET("/info", Route.GetBank)

	mobilePromotionGroup := mobileGroup.Group("/promotion")
	mobilePromotionGroup.GET("/info", Route.GetPromotion)

	cmsBranchGroup := cmsGroup.Group("/branch")
	cmsBranchGroup.POST("/create", Route.CreateBranch)
	cmsBranchGroup.PUT("/update", Route.UpdateBranch)
	cmsBranchGroup.DELETE("/delete", Route.DeleteBranch)
	cmsBranchGroup.POST("/info", Route.GetAllBranch)

	cmsPositionGroup := cmsGroup.Group("/position")
	cmsPositionGroup.POST("/create", Route.CreatePosition)
	cmsPositionGroup.PUT("/update", Route.UpdatePosition)
	cmsPositionGroup.DELETE("/delete", Route.DeletePosition)

	cmsPromotionGroup := cmsGroup.Group("/promotion")
	cmsPromotionGroup.POST("/create", Route.CreatePromotion)
	cmsPromotionGroup.PUT("/update", Route.UpdatePromotion)
	cmsPromotionGroup.DELETE("/delete", Route.DeletePromotion)

	cmsStaffGroup := cmsGroup.Group("/staff")
	cmsStaffGroup.POST("/create", Route.CreateStaff)
	cmsStaffGroup.PUT("/update", Route.UpdateStaff)
	cmsStaffGroup.DELETE("/delete", Route.DeleteStaff)
	cmsStaffGroup.POST("/prepare", Route.GetPrepareStaff)
	cmsStaffGroup.POST("/info", Route.GetAllStaff)

	cmsAddressGroup := cmsGroup.Group("/address")
	cmsAddressGroup.POST("/province", Route.GetProvince)
	cmsAddressGroup.POST("/amphur", Route.GetAmphur)
	cmsAddressGroup.POST("/district", Route.GetDistrict)

	e.Start(":8080")
}
