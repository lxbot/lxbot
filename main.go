package main

func main() {
	ch := loadAdapters()

	for {
		m := <- *ch

		println(m)
	}
}