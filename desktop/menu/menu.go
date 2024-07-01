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
	OnClick     func()

	// obj entity.Node
}

// func (i *Item) ComponentAttached(com entity.Node) {
// 	i.obj = entity.Parent(com)
// }

type Menu struct {
	Items []Item

	// obj entity.Node
	menu
}

func New(items []Item) *Menu {
	menu := &Menu{
		Items: items,
	}
	menu.Reload()
	return menu
}

// func (m *Menu) ComponentAttached(com entity.Node) {
// 	m.obj = entity.Parent(com)
// }

func (m *Menu) Popup() int {
	return m.popup()
}

func (m *Menu) Unload() {
	m.unload()
}

func (m *Menu) Reload() {
	m.unload()

	var items []Item
	if len(m.Items) > 0 {
		items = m.Items
	}
	// if len(m.Items) == 0 && m.obj != nil {
	// 	for _, i := range node.GetAll[*Item](m.obj, node.Include{Children: true, NotSelf: true}) {
	// 		items = append(items, *i)
	// 	}
	// }
	m.reload(items)
}

func subItems(item Item) (items []Item) {
	if len(item.SubMenu) > 0 {
		return item.SubMenu
	}
	// if item.obj != nil {
	// 	for _, i := range node.GetAll[*Item](item.obj, node.Include{Children: true, NotSelf: true}) {
	// 		items = append(items, *i)
	// 	}
	// }
	return
}
