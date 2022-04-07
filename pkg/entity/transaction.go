package entity

type Transaction struct {
	Hash    string      `json:"tx_hash"`
	From    string      `json:"from"`
	To      string      `json:"to"`
	Nonce   uint64      `json:"nonce"`
	Data    string      `json:"data"`
	Value   string      `json:"value"`
	LogList []*EventLog `json:"logs"`
}

type EventLog struct {
	Index uint   `json:"index"`
	Data  string `json:"data"`
}
