package startPkg

import (
	"log"
	"net/http"
	"os"

	"../broadcastHandle"

	"../account"
	"../adminModule"
	"../block"
	"../dashboard"
	"../errorPkgModule"
	file "../filestoragemodule"
	"../globalPkg"
	"../heartbeat"
	"../ledger"
	"../privacyandterms"
	"../serverworkload"
	"../serviceModule"
	"../tokenModule"
	"../transactionModule"
	"../validatorModule"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const path = "upload"

func HandleRequest() {
	myRouter := mux.NewRouter().StrictSlash(true)

	myRouter.HandleFunc("/a021d8007a2c590bc64ff2338d34c4e2", broadcastHandle.BroadcastHandle).Methods("POST", "OPTIONS")
	//terms
	myRouter.Handle("/1473545cba6723fcd5c", globalPkg.IsAdmine(privacyandterms.AddAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/3ee045371b56292368d", globalPkg.IsAdmine(privacyandterms.UpdateAPI)).Methods("PUT", "OPTIONS")
	myRouter.Handle("/66e3a121ecbfc8367a3", globalPkg.IsAdmine(privacyandterms.GetAllAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/ef6f1731754f477a38f", globalPkg.IsUser(privacyandterms.GetByIDAPI)).Methods("POST", "OPTIONS")
	//validator
	myRouter.Handle("/8bwvda1c62cb74a413a522d", globalPkg.IsAdmine(validatorModule.ValidatorAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/c67193278293f1f7052138x", globalPkg.IsAdmine(validatorModule.ValidatorAPI)).Methods("PUT", "OPTIONS")
	myRouter.Handle("/1f3f8159094bc4a28051f3", globalPkg.IsAdmine(validatorModule.ValidatorAPI)).Methods("DELETE", "OPTIONS")
	myRouter.Handle("/ld9e74a2a463b4306b2269", globalPkg.IsAdmine(validatorModule.GetAllValidatorAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/4bc7282209c047bd45e1", globalPkg.IsAdmine(validatorModule.BroadcastValidatorAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/bc5ddbbbe2466c0667c7feb", globalPkg.IsAdmine(validatorModule.BroadcastValidatorAPI)).Methods("PUT", "OPTIONS")
	myRouter.Handle("/aaf8cbdea4407fd2a2a7", globalPkg.IsAdmine(validatorModule.BroadcastValidatorAPI)).Methods("DELETE", "OPTIONS")
	myRouter.Handle("/s0e9cy8653ga014f7dcdr54", globalPkg.IsAdmine(validatorModule.DeactiveNode)).Methods("POST", "OPTIONS")
	myRouter.HandleFunc("/ConfirmedValidatorAPI", validatorModule.ConfirmedValidatorAPI).Methods("GET", "OPTIONS")
	myRouter.HandleFunc("/cfd32f1b77c4ac9b5e62749", validatorModule.GetnumberValidatorsAPI).Methods("POST", "OPTIONS")
	//account

	myRouter.Handle("/6247925023f757e98aecfd", globalPkg.IsAdmine(account.GetAllAccountsAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/66a12ba305638714a779r", globalPkg.IsUser(account.GetAccountInfoByAccountPublicKeyAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/b2fe578761b07d5b65bbb", globalPkg.IsAdmine(account.GetAllEmailsUsernameAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/b21a808878fb76a3ecfa56f", globalPkg.IsAdmine(account.GetnumberAccountsAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/GetAllpkAPI", globalPkg.IsAdmine(account.GetAllpkAPI)).Methods("POST", "OPTIONS")
	myRouter.HandleFunc("/b21a808878fb76a3ecfa56", account.GetnumAccountsAPI).Methods("POST", "OPTIONS")
	myRouter.HandleFunc("/9f77b63b695efde9d613606ce05d", account.GetAddressbyNameAPI).Methods("POST", "OPTIONS")
	//token
	myRouter.Handle("/4e5bdcfee9fcd710f4f7c7d", globalPkg.IsAdmine(tokenModule.GetAllTokenssAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/a34c260d1b9e1e0482b4e90", globalPkg.IsUser(tokenModule.RegisteringNewTokenAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/911b2f6bf2fef2f0633d60c7d", globalPkg.IsUser(tokenModule.UpdatingTokenAPI)).Methods("PUT", "OPTIONS")
	myRouter.Handle("/8iapf151296c8bdg0be5b8w", globalPkg.IsUser(tokenModule.ExploringUserTokensAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/512b2e814f397c47de6f5web", globalPkg.IsUser(tokenModule.RefundToken)).Methods("POST", "OPTIONS")
	myRouter.Handle("/tokenbyname", globalPkg.IsUser(tokenModule.GettokennameAPI)).Methods("POST", "OPTIONS")
	myRouter.HandleFunc("/1d3fdc392ce9d3c8c54e1ec", tokenModule.GetTokensValueLastdaysAPI).Methods("POST", "OPTIONS")
	
	//account module

	myRouter.HandleFunc("/9d6322c1f4d9d3f38aed8bf", account.Login).Methods("POST", "OPTIONS")                      //checked  solved
	myRouter.Handle("/06623be7673e8a781bffed9", globalPkg.IsUser(account.ChangeStatus)).Methods("POST", "OPTIONS") /////checked
	myRouter.HandleFunc("/833194c9bb5d54419e8cf7", account.UserRegister).Methods("POST", "OPTIONS")                //checked
	myRouter.Handle("/ed0a92ef7epa510862bu13", globalPkg.IsAdmine(account.ServiceRegisterAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/f45ca7354dcfbee245cdd62", globalPkg.IsAdmine(account.ServiceUpdateAPI)).Methods("PUT", "OPTIONS")
	myRouter.Handle("/99dc78047a8542fcf9decyp", globalPkg.IsUser(account.SavePublickey)).Methods("POST", "OPTIONS")
	myRouter.Handle("/28e773082adf7aa0e4f76e8", globalPkg.IsUser(account.UpdateAccountInfo)).Methods("PUT", "OPTIONS") //checked
	// myRouter.Handle("/819c7f06486287e0a6c25f00", globalPkg.IsAdmine(account.ConfirmatinByEmail)).Methods("GET", "OPTIONS")
	myRouter.HandleFunc("/12c38228a7c05c63f55cf66c", account.ConfirmationMessage).Methods("POST", "OPTIONS")            //checked solved
	myRouter.Handle("/6c954f53e30ceed92348ed9", globalPkg.IsUser(account.GetSearchProperty)).Methods("POST", "OPTIONS") //checked
	// myRouter.Handle("/cabaf46ee69d7b8445a5d791", globalPkg.IsUser(account.ForgetPassword)).Methods("POST", "OPTIONS")
	myRouter.HandleFunc("/cabaf46ee69d7b8445a5d791", account.ForgetPassword).Methods("POST", "OPTIONS")

	myRouter.Handle("/ef3d81592439a480ddea0647", globalPkg.IsUser(account.GetPkandValidatorPkUsingAddress)).Methods("POST", "OPTIONS") //checked
	//	myRouter.Handle("/819c7f06486287e0a6c25f00", account.ConfirmatinByEmail).Methods("GET", "OPTIONS")
	myRouter.HandleFunc("/{path}", account.ConfirmatinByEmail).Methods("GET", "OPTIONS") ///checked
	//transaction

	myRouter.Handle("/05a00dce22722e71dc8912", globalPkg.IsAdmine(transactionModule.GetAllTransactionsAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/2e4a9d667ad5e3cef02eae9", globalPkg.IsUser(transactionModule.AddNewTransaction)).Methods("POST", "OPTIONS")
	myRouter.Handle("/d1c22221807844b6515d806", globalPkg.IsUser(transactionModule.GetBalance)).Methods("POST", "OPTIONS")
	myRouter.Handle("/46cda9a3c810848b8141sfk", globalPkg.IsUser(transactionModule.GetTransactionByPublicKey)).Methods("POST", "OPTIONS")
	myRouter.Handle("/aa087e12bb63b56c295d7c8", globalPkg.IsAdmine(transactionModule.GetAllTransactionDbAPI)).Methods("POST", "OPTIONS")  //from db
	myRouter.Handle("/6b3ebe58321d8c0cb479bd0", globalPkg.IsAdmine(transactionModule.GetTransactionDbByIdAPI)).Methods("POST", "OPTIONS") //from db
	myRouter.Handle("/ced66eefdae6c69bcc17e5c11", globalPkg.IsUser(transactionModule.GetAllTransactionforOneTokenAPI)).Methods("POST", "OPTIONS")
	//Block

	myRouter.Handle("/783904af8918ebd13434d", globalPkg.IsAdmine(block.GetAllBlocksAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/64d380ea3fd7ac0ad9037", globalPkg.IsAdmine(block.GetBlockByIDAPI)).Methods("POST", "OPTIONS")
	// myRouter.HandleFunc("/addblock", block.AddBlockAPI).Methods("POST", "OPTIONS")
	//Ledger
	myRouter.HandleFunc("/221a80a3bc66c047be0fb", ledger.PostLegderAPI).Methods("POST", "OPTIONS")
	//myRouter.Handle("/2c3920b33633a95417ea", globalPkg.IsAdmine(ledger.GetLegderAPI)).Methods("POST", "OPTIONS")
	myRouter.HandleFunc("/2c3920b33633a95417ea", ledger.GetLegderAPI).Methods("POST", "OPTIONS")
	myRouter.Handle("/2c3720b33633a95417ea55", globalPkg.IsAdmine(ledger.GetTmpAccountDB)).Methods("POST", "OPTIONS")
	//error

	myRouter.Handle("/57877f5506b017626ce9", globalPkg.IsAdmine(errorPkgModule.GetAllErrorsAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/02ad27a1f7d1dd3da57e", globalPkg.IsAdmine(errorPkgModule.DeleteError)).Methods("POST", "OPTIONS")
	myRouter.Handle("/5w81f7d1dd3d01726cej", globalPkg.IsAdmine(errorPkgModule.DeleteErrorsBetweenTwoTimes)).Methods("POST", "OPTIONS")
	//proof of stake
	// myRouter.HandleFunc("/LotteryWinner", proof.LotteryWinner).Methods("POST", "OPTIONS")

	//global variables
	myRouter.Handle("/e402889ad34f58fca12f9", globalPkg.IsAdmine(globalPkg.PostGlobalVariableAPI)).Methods("POST", "OPTIONS")
	//heartbeat
	myRouter.Handle("/54f680b6e8476a8e4f5bedb3", globalPkg.IsAdmine(heartbeat.GetAllHeartBeat)).Methods("POST")
	//Dashboard
	myRouter.Handle("/faa902663bcf4b3cbae01ae", globalPkg.IsAdmine(dashboard.GetBlockDashboard)).Methods("POST")
	myRouter.Handle("/3547F1620102A5E6853B3B1", globalPkg.IsAdmine(dashboard.GetStatistics)).Methods("POST")
	//serverworkload
	myRouter.Handle("/4b5af442229cf356a686", globalPkg.IsAdmine(serverworkload.AllStats)).Methods("POST", "OPTIONS")
	myRouter.Handle("/ee76c04de65e414fcdae", globalPkg.IsAdmine(serverworkload.CPUStats)).Methods("POST", "OPTIONS")
	myRouter.Handle("/bc6071868954e08066b", globalPkg.IsAdmine(serverworkload.MemoryStats)).Methods("POST", "OPTIONS")
	myRouter.Handle("/3dab660431d6f68e0b5", globalPkg.IsAdmine(serverworkload.StorageStats)).Methods("POST", "OPTIONS")
	myRouter.Handle("/f32b5ae8ade89322fb5s", globalPkg.IsAdmine(serverworkload.NetworkStats)).Methods("POST", "OPTIONS")
	myRouter.Handle("/92c4504e726459a357d", globalPkg.IsAdmine(serverworkload.CurrentTransactionsCount)).Methods("POST", "OPTIONS")

	// Admin
	myRouter.Handle("/090cafbf09ef29c167d019", globalPkg.IsAdmine(adminModule.GetAllAdminsAPI)).Methods("POST", "OPTIONS")
	//myRouter.HandleFunc("/a3343a4fd0e0abf6d05d21", adminModule.GetAdminByIDAPI).Methods("POST", "OPTIONS")
	myRouter.Handle("/82ab2a0f31306fc9b5a25f", globalPkg.IsAdmine(adminModule.AddNewAdmin)).Methods("POST", "OPTIONS")
	// myRouter.Handle("/597e8858b6b2d9d0373f0", globalPkg.IsAdmine(adminModule.LoginAdmin)).Methods("POST", "OPTIONS")
	myRouter.HandleFunc("/597e8858b6b2d9d0373f0", adminModule.LoginAdmin).Methods("POST", "OPTIONS")
	myRouter.Handle("/e7ec238009cf94c49a89a9", globalPkg.IsAdmine(adminModule.UpdateAdmin)).Methods("PUT", "OPTIONS")
	myRouter.Handle("/514e4o5cbcfay4eb72b470a4", globalPkg.IsAdmine(adminModule.GetAlltransactionPerMonthAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/ef48420r204e05bfgi5s8e1y", globalPkg.IsAdmine(adminModule.GetAlltransactionLastTenMinuteAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/a1c8b6ab207051c3350aabe8a", globalPkg.IsAdmine(adminModule.UpdatesuperAdmin)).Methods("PUT", "OPTIONS")
	myRouter.Handle("/f98f386b9d679815a13c7cdedb", globalPkg.IsAdmine(adminModule.GetadminsofSuperAdminAPI)).Methods("POST", "OPTIONS")
	
	myRouter.Handle("/b512e6d3fb3e7015880d275b", globalPkg.IsUser(serviceModule.InquiryNewInternetServiceCost)).Methods("POST", "OPTIONS")
	myRouter.Handle("/29545d292781d573475ad68c", globalPkg.IsUser(serviceModule.PurchaseService)).Methods("POST", "OPTIONS")
	myRouter.Handle("/1acbf975e4f59bffae27ce43", globalPkg.IsUser(serviceModule.GetAllPurchasedServices)).Methods("POST", "OPTIONS")
	myRouter.Handle("/b64c260d1b9e1e044f397c50", globalPkg.IsUser(serviceModule.CheckVoucherStatus)).Methods("POST", "OPTIONS")
	myRouter.Handle("/1acbf975e4f59bffae27ce42", globalPkg.IsUser(serviceModule.GetAllNamesandPKsForServiceAccount)).Methods("POST", "OPTIONS")

	//UploadFile
	myRouter.Handle("/a75a09dd2f8d64e6e7f171bea03033e4c00c1c", globalPkg.IsAdmine(file.GetAllChunksAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/fe6b561cf1187e106e3dfb6829b2e12c78783", globalPkg.IsUser(file.UploadFile)).Methods("POST", "OPTIONS")
	myRouter.Handle("/a9f0be2e898c28ac59c79d642722b88966c", globalPkg.IsUser(file.ExploreFiles)).Methods("POST", "OPTIONS")
	myRouter.Handle("/bbfcf281d1c9c237c81e7b5526280", globalPkg.IsUser(file.DeleteFile)).Methods("POST", "OPTIONS")
	myRouter.Handle("/abb3e140af96a1dfd63803e716473c", globalPkg.IsUser(file.RetrieveFile)).Methods("POST", "OPTIONS")
	myRouter.Handle("/940b14d596a4ec42619412ff0dccb", globalPkg.IsAdmine(file.GetAllShareFileAPI)).Methods("POST", "OPTIONS")
	myRouter.Handle("/c4f5cf2170b878cfeb691eeb8a795", globalPkg.IsUser(file.ShareFiles)).Methods("POST", "OPTIONS")
	myRouter.Handle("/fbd4f20dfe6b57e8979c7bb41bc", globalPkg.IsUser(file.UnshareFile)).Methods("POST", "OPTIONS")
	// make a file server where upload files
	os.Mkdir(path, 0777)
	myRouter.PathPrefix("/files/").Handler(http.StripPrefix("/files/", http.FileServer(http.Dir(path)))).Methods("GET", "OPTIONS")
	//Reset password
	myRouter.HandleFunc("/{Path}", account.ResetPassword).Methods("PUT", "OPTIONS")
	// function to start go-routines
	StartGoRoutine()
	log.Fatal(http.ListenAndServe(":"+Conf.Server.Port, handlers.CORS(handlers.AllowedHeaders([]string{"X-Requested-With", "token", "Content-Type", "Authorization", "jwt-token"}), handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS"}), handlers.AllowedOrigins([]string{"*"}))(myRouter)))
}
