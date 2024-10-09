package main

import (
	"fmt"
	"io"
	"os"

	"tractor.dev/toolkit-go/desktop"
	"tractor.dev/toolkit-go/desktop/app"
	"tractor.dev/toolkit-go/desktop/menu"
	"tractor.dev/toolkit-go/desktop/window"
)

var w *window.Window

func read_png_to_byte_array(file_path string) ([]byte, error) {
	// Open the file
	file, err := os.Open(file_path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the file into a byte array
	file_bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return file_bytes, nil
}

func main() {
	file_path := "icon.png"
	byte_array, err := read_png_to_byte_array(file_path)
	if err != nil {
		fmt.Printf("Error reading PNG file: %v\n", err)
		return
	}

	desktop.Start(func() {
		app.Run(app.Options{}, func() {
			w := window.New(window.Options{
				URL:   "https://google.com",
				Title: "Hello",
				Size: desktop.Size{
					Width:  1000,
					Height: 1000,
				},
				Visible:     true,
				Resizable:   true,
				Center:      true,
				Transparent: true,
			})
			w.Reload()
			w.SetTitle("booo")
			items := []menu.Item{
				{
					ID:        0,
					Title:     "Title1",
					Disabled:  false,
					Selected:  false,
					Separator: false,
				},
				{
					ID:        1,
					Title:     "",
					Disabled:  false,
					Selected:  false,
					Separator: true,
				},
				{
					ID:        2,
					Title:     "Title3",
					Disabled:  false,
					Selected:  false,
					Separator: false,
					SubMenu: []menu.Item{
						{
							ID:        21,
							Title:     "Title3.1",
							Disabled:  false,
							Selected:  true,
							Separator: false,
						},
					},
				},
			}
			app.NewIndicator(byte_array, items)
		})

	})

}