package pdf

import (
	"context"
	"io"

	"github.com/go-swiss/fonts"
	"github.com/phpdave11/gofpdf"
)

type FpdfConfig struct {
	Title   string
	Subject string

	Orientation string // Default Portrait
	PaperSize   string // Default A4

	Logo       io.Reader
	LogoFormat string

	// Header
	HeaderFunc func(impl Fpdf, fontsCache fonts.Cache) func()

	// Footer
	FooterFunc func(impl Fpdf, fontsCache fonts.Cache) func()
}

func NewFpdf(ctx context.Context, c FpdfConfig, fontsCache fonts.Cache) *Fpdf {
	fpdf := Fpdf{
		Fpdf:    gofpdf.New(c.Orientation, "pt", c.PaperSize, ""),
		anchors: make(map[string]int),
	}

	fpdf.Fpdf.SetTitle(c.Title, true)
	fpdf.Fpdf.SetSubject(c.Subject, true)
	fpdf.Fpdf.SetCellMargin(0)

	if c.HeaderFunc != nil {
		fpdf.Fpdf.SetHeaderFunc(c.HeaderFunc(fpdf, fontsCache))
	}

	if c.FooterFunc != nil {
		fpdf.Fpdf.SetFooterFunc(c.FooterFunc(fpdf, fontsCache))
	}

	fpdf.AddPage()

	return &fpdf
}

type internalLink struct {
	page          int
	x, y          float64
	width, height float64
	anchor        string
}

type Fpdf struct {
	Fpdf        *gofpdf.Fpdf
	anchorLinks []internalLink
	anchors     map[string]int
}

// Add a new page
func (f Fpdf) AddPage() {
	f.Fpdf.AddPage()
}

// Position
func (f Fpdf) GetX() float64 {
	return f.Fpdf.GetX()
}

func (f Fpdf) GetY() float64 {
	return f.Fpdf.GetY()
}

func (f Fpdf) SetX(x float64) {
	f.Fpdf.SetX(x)
}

func (f Fpdf) SetY(y float64) {
	f.Fpdf.SetY(y)
}

// Page size
func (f Fpdf) GetPageSize() (width float64, height float64) {
	return f.Fpdf.GetPageSize()
}

// Margins
func (f Fpdf) SetMarginLeft(margin float64) {
	f.Fpdf.SetLeftMargin(margin)
}

func (f Fpdf) SetMarginRight(margin float64) {
	f.Fpdf.SetRightMargin(margin)
}

func (f Fpdf) SetMarginTop(margin float64) {
	f.Fpdf.SetTopMargin(margin)
}

func (f Fpdf) SetFont(family string, style string, size int) error {
	f.Fpdf.SetFont(family, style, float64(size))
	return nil
}

// Writing
func (f Fpdf) WriteText(height float64, text string) {
	f.Fpdf.Write(height, text)
}

func (f Fpdf) CellFormat(w float64, h float64, txtStr string, borderStr string, ln int, alignStr string, fill bool, link int, linkStr string) {
	f.Fpdf.SetCellMargin(0)
	f.Fpdf.CellFormat(w, h, txtStr, borderStr, ln, alignStr, fill, link, linkStr)
}

func (f *Fpdf) AddInternalLink(anchor string) {
	linkID := f.Fpdf.AddLink()
	f.Fpdf.SetLink(linkID, f.GetY(), -1)
	f.anchors[anchor] = linkID
}

func (f *Fpdf) WriteInternalLink(lineHeight float64, text string, anchor string) {
	f.anchorLinks = append(f.anchorLinks, internalLink{
		page:   f.Fpdf.PageNo(),
		width:  f.MeasureTextWidth(text),
		height: lineHeight,
		x:      f.GetX(),
		y:      f.GetY(),
		anchor: anchor,
	})
	f.Fpdf.WriteLinkString(lineHeight, text, "#"+anchor)
}

func (f Fpdf) WriteExternalLink(lineHeight float64, text string, destination string) {
	f.Fpdf.WriteLinkString(lineHeight, text, destination)
}

func (f Fpdf) BR(height float64) {
	f.Fpdf.Ln(height)
}

// Images
func (f Fpdf) RegisterImage(id string, format string, src io.Reader) {
	f.Fpdf.RegisterImageOptionsReader(id, gofpdf.ImageOptions{ImageType: format, ReadDpi: false}, src)
}

func (f Fpdf) UseImage(imgID string, x, y, w, h float64) {
	f.Fpdf.ImageOptions(imgID, x, y, w, h, true, gofpdf.ImageOptions{ImageType: "", ReadDpi: false}, 0, "")
}

// Measuring
func (f Fpdf) MeasureTextWidth(text string) float64 {
	return f.Fpdf.GetStringWidth(text)
}

func (f Fpdf) SplitText(txt string, w float64) []string {
	lines := f.Fpdf.SplitLines([]byte(txt), w)

	split := make([]string, len(lines))
	for k, line := range lines {
		split[k] = string(line)
	}

	return split
}

// Colors
func (f Fpdf) SetDrawColor(r uint8, g uint8, b uint8) {
	f.Fpdf.SetDrawColor(int(r), int(g), int(b))
}

func (f Fpdf) SetFillColor(r uint8, g uint8, b uint8) {
	f.Fpdf.SetFillColor(int(r), int(g), int(b))
}

func (f Fpdf) SetTextColor(r uint8, g uint8, b uint8) {
	f.Fpdf.SetTextColor(int(r), int(g), int(b))
}

// Width
func (f Fpdf) SetLineWidth(width float64) {
	f.Fpdf.SetLineWidth(width)
}

func (f Fpdf) Line(x1 float64, y1 float64, x2 float64, y2 float64) {
	f.Fpdf.MoveTo(x1, y1)
	f.Fpdf.LineTo(x2, y2)
	f.Fpdf.DrawPath("F")
}

func (f Fpdf) Write(w io.Writer) error {
	// write the internal links
	for _, link := range f.anchorLinks {
		id, ok := f.anchors[link.anchor]
		if !ok {
			continue
		}

		f.Fpdf.SetPage(link.page)
		f.Fpdf.Link(link.x, link.y, link.width, link.height, id)
	}

	f.Fpdf.SetPage(f.Fpdf.PageCount())
	return f.Fpdf.Output(w)
}

func (f Fpdf) GetMargins() (left, top, right, bottom float64) {
	return f.Fpdf.GetMargins()
}

func (f Fpdf) AddFont(family string, style string, data []byte) error {
	f.Fpdf.AddUTF8FontFromBytes(family, style, data)
	return nil
}
