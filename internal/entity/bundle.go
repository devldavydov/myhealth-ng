package entity

type BundleTotals struct {
	Weight  float64
	Cal     float64
	Protein float64
	Fat     float64
	Carb    float64
}

type BundleItem struct {
	Food   Food
	Weight float64
}

type BundleItemData struct {
	FoodKey string
	Weight  float64
}

type Bundle struct {
	Key    string
	Name   string
	Items  []BundleItem
	Totals BundleTotals
}

type BundleSummary struct {
	Key       string
	Name      string
	ItemCount int
	Totals    BundleTotals
}

type BundleData struct {
	Name  string
	Items []BundleItemData
}
