package devopsruntime

func GetNvmeCount() (int, error) {
	return len(GetNvmeList()), nil
}

func GetNvmeDesc() ([]string, error) {
	nvmes := []string{}

	for _, nvme := range GetNvmeList() {
		nvmes = append(nvmes, Info2String(nvme))
	}

	return nvmes, nil
}

func GetGpuCount() (int, error) {
	return len(GetGpuList()), nil
}

func GetGpuDesc() ([]string, error) {
	gpus := []string{}

	for _, gpu := range GetGpuList() {
		gpus = append(gpus, Info2String(gpu))
	}

	return gpus, nil
}

func GetMemoryCount() (int, error) {
	return len(GetMemoryList()), nil
}

func GetMemorySize() (uint64, error) {
	var memSizeGB uint64 = 0

	for _, mem := range GetMemoryList() {
		memSizeGB += uint64(mem.SizeGB)
	}

	return memSizeGB * 1024 * 1024 * 1024, nil
}

func GetMemoryDesc() ([]string, error) {
	mems := []string{}

	for _, mem := range GetMemoryList() {
		mems = append(mems, Info2String(mem))
	}

	return mems, nil
}

func GetCpuCount() (int, error) {
	return len(GetCpuList()), nil
}

func GetCpuDesc() ([]string, error) {
	cpus := []string{}

	for _, cpu := range GetCpuList() {
		cpus = append(cpus, Info2String(cpu))
	}

	return cpus, nil
}

func GetHddCount() (int, error) {
	return len(GetHddList()), nil
}

func GetHddDesc() ([]string, error) {
	hdds := []string{}

	for _, hdd := range GetHddList() {
		hdds = append(hdds, Info2String(hdd))
	}

	return hdds, nil
}

func GetEthernetCount() (int, error) {
	return len(GetEthernetList()), nil
}

func GetEthernetDesc() ([]string, error) {
	eths := []string{}

	for _, eth := range GetEthernetList() {
		eths = append(eths, Info2String(eth))
	}

	return eths, nil
}
