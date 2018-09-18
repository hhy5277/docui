package panel

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/jroimartin/gocui"
)

var imageNameRegexp = regexp.MustCompile("\\[.*\\]")

type ImageList struct {
	*Gui
	name string
	Position
}

func NewImageList(gui *Gui, name string, x, y, w, h int) ImageList {
	return ImageList{gui, name, Position{x, y, x + w, y + h}}
}

func (i ImageList) Name() string {
	return i.name
}

func (i ImageList) SetView(g *gocui.Gui) (*gocui.View, error) {
	v, err := g.SetView(i.Name(), i.x, i.y, i.w, i.h)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return nil, err
		}

		v.Title = v.Name()
		v.Wrap = true

		if _, err = SetCurrentPanel(g, i.Name()); err != nil {
			return nil, err
		}

		return v, nil
	}

	return v, nil
}

func (i ImageList) Init(g *Gui) {
	v, err := i.SetView(g.Gui)

	if err != nil {
		panic(err)
	}

	i.LoadImages(v)
	v.SetCursor(0, 1)

	// keybinds
	g.SetKeybinds(i.Name())

	if err := g.SetKeybinding(i.Name(), Key("j"), gocui.ModNone, CursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), Key("k"), gocui.ModNone, CursorUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), gocui.KeyEnter, gocui.ModNone, i.DetailImage); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), Key("o"), gocui.ModNone, i.DetailImage); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), Key("c"), gocui.ModNone, i.CreateContainerPanel); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), Key("p"), gocui.ModNone, i.PullImagePanel); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), Key("d"), gocui.ModNone, i.RemoveImage); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), Key("e"), gocui.ModNone, i.ExportImage); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), Key("i"), gocui.ModNone, i.ImportImage); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(i.Name(), gocui.KeyCtrlL, gocui.ModNone, i.LoadImage); err != nil {
		log.Panicln(err)
	}
}

func (i ImageList) CreateContainerPanel(g *gocui.Gui, v *gocui.View) error {
	id := i.GetImageID(v)
	if id == "" {
		return nil
	}

	data := map[string]interface{}{
		"Image": id,
	}

	maxX, maxY := i.Size()
	x := maxX / 8
	y := maxY / 8
	w := maxX - x
	h := maxY - y
	input := NewInput(i.Gui, CreateContainerPanel, x, y, w, h, NewCreateContainerItems(x, y, w, h), data)
	input.Init(i.Gui)
	return nil
}

func (i ImageList) PullImagePanel(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := i.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := y + 4

	input := NewInput(i.Gui, PullImagePanel, x, y, w, h, NewPullImageItems(x, y, w, h), make(map[string]interface{}))
	input.Init(i.Gui)
	return nil

}

func (i ImageList) DetailImage(g *gocui.Gui, v *gocui.View) error {

	id := i.GetImageID(v)
	if id == "" {
		return nil
	}

	img, err := i.Docker.InspectImage(id)
	if err != nil {
		return err
	}

	nv, err := g.View(DetailPanel)
	if err != nil {
		panic(err)
	}

	nv.Clear()
	nv.SetOrigin(0, 0)
	nv.SetCursor(0, 0)
	fmt.Fprint(nv, StructToJson(img))

	return nil
}

func (i ImageList) ExportImage(g *gocui.Gui, v *gocui.View) error {

	id := i.GetImageName(v)
	if id == "" {
		return nil
	}

	maxX, maxY := i.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := y + 4

	data := map[string]interface{}{
		"ID": id,
	}

	input := NewInput(i.Gui, ExportImagePanel, x, y, w, h, NewExportImageItems(x, y, w, h), data)
	input.Init(i.Gui)
	return nil
}

func (i ImageList) ImportImage(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := i.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := maxY - y

	input := NewInput(i.Gui, ImportImagePanel, x, y, w, h, NewImportImageItems(x, y, w, h), make(map[string]interface{}))
	input.Init(i.Gui)
	return nil
}

func (i ImageList) LoadImage(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := i.Size()
	x := maxX / 3
	y := maxY / 3
	w := maxX - x
	h := y + 4

	input := NewInput(i.Gui, LoadImagePanel, x, y, w, h, NewLoadImageItems(x, y, w, h), make(map[string]interface{}))
	input.Init(i.Gui)

	return nil
}

func (i ImageList) RefreshPanel(g *gocui.Gui, v *gocui.View) error {
	if v == nil {
		nv, err := g.View(ImageListPanel)
		if err != nil {
			return err
		}

		v = nv
	}
	v.Clear()
	i.LoadImages(v)
	SetCurrentPanel(g, v.Name())
	return nil
}

func (i ImageList) LoadImages(v *gocui.View) {
	fmt.Fprintf(v, "%-15s %-20s\n", "ID", "NAME")
	for _, i := range i.Docker.Images() {
		fmt.Fprintf(v, "%-15s %-20s\n", i.ID[7:19], i.RepoTags)
	}
}

func (i ImageList) GetImageID(v *gocui.View) string {
	id := ReadLine(v, nil)
	if id == "" || id[:2] == "ID" {
		return ""
	}

	return id[:12]
}

func (i ImageList) GetImageName(v *gocui.View) string {
	line := ReadLine(v, nil)
	if line == "" || line[:2] == "ID" {
		return ""
	}

	name := imageNameRegexp.FindAllStringSubmatch(line, -1)[0][0]
	return strings.TrimRight(name[1:len(name)-1], " ")
}

func (i ImageList) RemoveImage(g *gocui.Gui, v *gocui.View) error {
	i.PrePanel = ImageListPanel
	name := i.GetImageID(v)
	if name == "" {
		return nil
	}

	i.ConfirmMessage("Do you want delete this image? (y/n)", func(g *gocui.Gui, v *gocui.View) error {
		if err := i.Docker.RemoveImageWithName(name); err != nil {
			i.CloseConfirmMessage(g, v)
			i.DispMessage(err.Error(), ImageListPanel)
			return nil
		}

		i.CloseMessage(g, v)

		return nil
	})

	return nil
}