package menu

var (
	mainMenu *Menu
)

func Main() *Menu {
	return mainMenu
}

func SetMain(menu *Menu) error {
	mainMenu = menu
	return nil
}

type Item struct {
	ID          int
	Title       string
	Disabled    bool
	Selected    bool
	Separator   bool
	Accelerator string
	SubMenu     []Item
	OnExecute   func()
}

func (i *Item) SubItems() (items []Item) {
	if len(i.SubMenu) > 0 {
		return i.SubMenu
	}
	return
}

func (i *Item) Execute() {
	if i.OnExecute != nil {
		i.OnExecute()
	}
}

type Menu struct {
	Items []Item

	menu
}

func New(items []Item) *Menu {
	menu := &Menu{
		Items: items,
	}
	for _, it := range items {
		menu.AddItem(it)
	}
	menu.Reload()
	return menu
}

func (m *Menu) Popup() int {
	return m.popup()
}

func (m *Menu) Unload() {
	m.unload()
}

func (m *Menu) Reload() {
	m.unload()
	m.load()

	// if len(m.Items) == 0 && m.obj != nil {
	// 	for _, i := range node.GetAll[*Item](m.obj, node.Include{Children: true, NotSelf: true}) {
	// 		items = append(items, *i)
	// 	}
	// }

}
