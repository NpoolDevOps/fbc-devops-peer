package minerapi

type MinerInfo struct {
	Power    float64
	Raw      float64
	Commited float64
	Proving  float64
	Faulty   float64

	MinerBalance     float64
	InitialPledge    float64
	PrecommitDeposit float64
	Vesting          float64
	Available        float64

	WorkerBalance  float64
	ControlBalance float64

	State map[string]uint64
}

func GetMinerInfo() *MinerInfo {

}
