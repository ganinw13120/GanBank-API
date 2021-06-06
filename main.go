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
	mobileCustomerGroup.POST("/has", Route.HasCustomer)
	mobileCustomerGroup.POST("/haskey", Route.HasCustomerKey)
	mobileCustomerGroup.POST("/createkey", Route.CreateCustomerKey)
	mobileCustomerGroup.POST("/createsession", Route.CreateCustomerSession)
	mobileCustomerGroup.POST("/signout", Route.SignoutCustomerSession)
	mobileCustomerGroup.POST("/hassession", Route.HasCustomerSession)
	mobileCustomerGroup.POST("/qrcode", Route.GetQrcode)

	cmsAccountGroup := cmsGroup.Group("/account")
	cmsAccountGroup.POST("/create", Route.CreateAccount)
	cmsAccountGroup.POST("/info", Route.GetAccount)
	cmsAccountGroup.POST("/prepare", Route.PrepareAccount)
	cmsAccountGroup.POST("/name", Route.GetAccountName)
	cmsAccountGroup.POST("/suspend", Route.SuspendAccount)
	cmsAccountGroup.POST("/edit", Route.EditAccount)

	mobileAccountGroup := mobileGroup.Group("/account")
	mobileAccountGroup.POST("/info", Route.GetAccountByID)
	mobileAccountGroup.POST("/infobycustomer", Route.GetAccountByCustomer)
	mobileAccountGroup.POST("/transaction", Route.PrepareTransaction)
	mobileAccountGroup.POST("/transfer", Route.Transfer)
	mobileAccountGroup.POST("/infobyqrcode", Route.GetInfoByQrcode)
	mobileAccountGroup.POST("/lasttransfer", Route.GetLastedTransfer)

	mobileBankGroup := mobileGroup.Group("/bank")
	mobileBankGroup.GET("/info", Route.GetBank)

	mobilePromotionGroup := mobileGroup.Group("/promotion")
	mobilePromotionGroup.GET("/info", Route.GetPromotion)

	cmsBranchGroup := cmsGroup.Group("/branch")
	cmsBranchGroup.POST("/create", Route.CreateBranch)
	cmsBranchGroup.POST("/edit", Route.EditBranch)
	cmsBranchGroup.PUT("/update", Route.UpdateBranch)
	cmsBranchGroup.DELETE("/delete", Route.DeleteBranch)
	cmsBranchGroup.POST("/info", Route.GetAllBranch)
	cmsBranchGroup.POST("/deleteid", Route.DeleteBranchByID)

	cmsPositionGroup := cmsGroup.Group("/position")
	cmsPositionGroup.POST("/create", Route.CreatePosition)
	cmsPositionGroup.PUT("/update", Route.UpdatePosition)
	cmsPositionGroup.DELETE("/delete", Route.DeletePosition)

	cmsPromotionGroup := cmsGroup.Group("/promotion")
	cmsPromotionGroup.POST("/create", Route.CreatePromotion)
	cmsPromotionGroup.PUT("/update", Route.UpdatePromotion)
	cmsPromotionGroup.POST("/delete", Route.DeletePromotion)
	cmsPromotionGroup.POST("/get", Route.GetPromotion)
	cmsPromotionGroup.POST("/getbyid", Route.GetPromotionByID)

	cmsStaffGroup := cmsGroup.Group("/staff")
	cmsStaffGroup.POST("/create", Route.CreateStaff)
	cmsStaffGroup.POST("/edit", Route.EditStaff)
	cmsStaffGroup.PUT("/update", Route.UpdateStaff)
	cmsStaffGroup.POST("/delete", Route.DeleteStaff)
	cmsStaffGroup.POST("/prepare", Route.GetPrepareStaff)
	cmsStaffGroup.POST("/info", Route.GetAllStaff)

	cmsAddressGroup := cmsGroup.Group("/address")
	cmsAddressGroup.POST("/province", Route.GetProvince)
	cmsAddressGroup.POST("/amphur", Route.GetAmphur)
	cmsAddressGroup.POST("/district", Route.GetDistrict)

	cmsTransactionGroup := cmsGroup.Group("/transaction")
	cmsTransactionGroup.POST("/create", Route.CreateTransaction)
	cmsTransactionGroup.POST("/transfer", Route.CreateTransferTransaction)
	cmsTransactionGroup.POST("/info", Route.GetAllTransaction)
	cmsTransactionGroup.POST("/prepare", Route.GetAllBank)

	cmsLoanGroup := cmsGroup.Group("/loan")
	cmsLoanGroup.POST("/prepare", Route.GetPrepareLoan)
	cmsLoanGroup.POST("/create", Route.CreateLoan)
	cmsLoanGroup.POST("/info", Route.GetAllLoan)
	cmsLoanGroup.POST("/find", Route.GetLoanByID)
	cmsLoanGroup.POST("/update", Route.UpdateLoanStatus)

	cmsAuthGroup := cmsGroup.Group("/auth")
	cmsAuthGroup.POST("/login", Route.Login)
	cmsAuthGroup.POST("/check", Route.CheckSession)

	cmsCustomerGroup := cmsGroup.Group("/customer")
	cmsCustomerGroup.POST("/info", Route.GetAllCustomer)

	e.Start(":8080")
}
