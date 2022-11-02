package main

func main() {
	//connStr := "root:atk_2018@tcp(127.0.0.1:3306)/container-dev?charset=utf8mb4&parseTime=True&loc=Local"
	//db, err := gorm.Open(mysql.Open(connStr), &gorm.Config{})
	//if err != nil {
	//	panic(err)
	//}
	//
	//userInsert := dto.NewUserInput{
	//	Username:    "joexu01",
	//	RawPassword: "12345678",
	//	Email:       "joexu01@yahoo.com",
	//}
	//
	//pwdHash, err := public.GeneratePwdHash([]byte(userInsert.RawPassword))
	//if err != nil {
	//	panic(err)
	//}
	//user := dao.User{
	//	Id:        0,
	//	Username:  userInsert.Username,
	//	Password:  pwdHash,
	//	Email:     userInsert.Email,
	//	CreatedAt: time.Now(),
	//	UpdatedAt: time.Now(),
	//	IsDelete:  0,
	//}
	//
	//result := db.Create(&user)
	//
	//fmt.Println(user.Id)
	//
	//fmt.Printf("%+v", result)
	//
	//user1 := dao.User{
	//	Id:        0,
	//	Username:  "joexu01",
	//	Password:  "",
	//	Email:     "",
	//	CreatedAt: time.Time{},
	//	UpdatedAt: time.Time{},
	//	IsDelete:  0,
	//}
	//
	//result = db.First(&user1)
	//
	//fmt.Println(user1.Email)
	//fmt.Printf("%+v", result)

	//fmt.Printf("%s", string("hello\n\n\n"))
	//
	//ret := nvml.Init()
	//if ret != nvml.SUCCESS {
	//	log.Fatalf("Unable to initialize NVML: %v", nvml.ErrorString(ret))
	//}
	//defer func() {
	//	ret := nvml.Shutdown()
	//	if ret != nvml.SUCCESS {
	//		log.Fatalf("Unable to shutdown NVML: %v", nvml.ErrorString(ret))
	//	}
	//}()
	//
	//count, ret := nvml.DeviceGetCount()
	//if ret != nvml.SUCCESS {
	//	log.Fatalf("Unable to get device count: %v", nvml.ErrorString(ret))
	//}
	//
	//for i := 0; i < count; i++ {
	//	device, ret := nvml.DeviceGetHandleByIndex(i)
	//	if ret != nvml.SUCCESS {
	//		log.Fatalf("Unable to get device at index %d: %v", i, nvml.ErrorString(ret))
	//	}
	//
	//	uuid, ret := device.GetUUID()
	//	if ret != nvml.SUCCESS {
	//		log.Fatalf("Unable to get uuid of device at index %d: %v", i, nvml.ErrorString(ret))
	//	}
	//	fmt.Printf("GPU UUID: %v\n", uuid)
	//
	//	name, ret := device.GetName()
	//	if ret != nvml.SUCCESS {
	//		log.Fatalf("Unable to get name of device at index %d: %v", i, nvml.ErrorString(ret))
	//	}
	//	fmt.Printf("GPU Name: %+v\n", name)
	//
	//	memoryInfo, _ := device.GetMemoryInfo()
	//	fmt.Printf("Memory Info: %+v\n", memoryInfo)
	//
	//	powerUsage, _ := device.GetPowerUsage()
	//	fmt.Printf("Power Usage: %+v\n", powerUsage)
	//
	//	powerState, _ := device.GetPowerState()
	//	fmt.Printf("Power State: %+v\n", powerState)
	//
	//	managementDefaultLimit, _ := device.GetPowerManagementDefaultLimit()
	//	fmt.Printf("Power Managment Default Limit: %+v\n", managementDefaultLimit)
	//
	//	version, _ := device.GetInforomImageVersion()
	//	fmt.Printf("Info Image Version: %+v\n", version)
	//
	//	driverVersion, _ := nvml.SystemGetDriverVersion()
	//	fmt.Printf("Driver Version: %+v\n", driverVersion)
	//
	//	cudaDriverVersion, _ := nvml.SystemGetCudaDriverVersion()
	//	fmt.Printf("CUDA Driver Version: %+v\n", cudaDriverVersion)
	//
	//	computeRunningProcesses, _ := device.GetGraphicsRunningProcesses()
	//	for _, proc := range computeRunningProcesses {
	//		fmt.Printf("Proc: %+v\n", proc)
	//	}
	//}
	//
	//fmt.Println()
	//
	//proc, _ := process.NewProcess(386485)
	//name, _ := proc.Cmdline()
	//fmt.Printf("%+v   |   %s\n", proc, name)
	//
	//counts, _ := cpu.Counts(false)
	//fmt.Println("counts ", counts)
	//info, _ := cpu.Info()
	//
	//for idx := 0; idx < len(info); idx++ {
	//	//fmt.Printf("%+v\n", info[idx].ModelName)
	//	//fmt.Println("cores", info[idx].Cores)
	//	//fmt.Println(info[idx].VendorID)
	//	//fmt.Println(info[idx].Mhz)
	//	//fmt.Println(info[idx].Family)
	//	fmt.Println("core id", info[idx].CoreID)
	//	fmt.Println("phy id", info[idx].PhysicalID)
	//	fmt.Println()
	//}

	//u := &dao.ContainerUser{}
	//info, err := u.GetContainerList(&gin.Context{}, db, &dto.UserContainerListQueryInput{
	//	PageNo:   1,
	//	PageSize: 20,
	//})
	//if err != nil {
	//	log.Fatal("%+v\n\n", err)
	//}
	//
	//fmt.Printf("%+v\n", info)

	//fileInfos, err := ioutil.ReadDir("/home/joseph/repo")
	//if err != nil {
	//	log.Fatal(err.Error())
	//}
	//
	//for idx, f := range fileInfos {
	//	fmt.Printf("%d: %+v \n\n", idx, f)
	//}

}
