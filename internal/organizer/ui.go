package organizer

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
)

var categoryIcons = map[string]string{
	"images": "🖼 ", "videos": "🎬", "audio": "🎵", "documents": "📚",
	"archives": "🗜 ", "packages": "📦", "code": "💻", "config": "⚙ ",
	"scripts": "📜", "devops": "🛠 ", "blockchain": "⛓ ", "data": "📊",
	"design": "🎨", "temp": "🗑 ", "others": "📎",
}

func iconForCategory(category string) string {
	if icon, ok := categoryIcons[category]; ok {
		return icon
	}
	return categoryIcons["others"]
}

// reporter renders per-item lines and, on a terminal, a live progress bar
// pinned to the last line. Every line is truncated to the terminal width, so
// carriage-return redraws never wrap.
type reporter struct {
	out    io.Writer
	live   bool
	fd     int
	total  int
	done   int
	moved  int
	skip   int
	dryRun bool
	start  time.Time
	counts map[string]int

	dim, ok, warn, accent, name, head lipgloss.Style
}

func newReporter(out io.Writer, total int, dryRun bool) *reporter {
	r := &reporter{out: out, total: total, dryRun: dryRun, start: time.Now(), counts: map[string]int{}}
	if f, isFile := out.(*os.File); isFile && term.IsTerminal(int(f.Fd())) {
		r.fd = int(f.Fd())
		r.live = !dryRun
	}
	rend := lipgloss.NewRenderer(out)
	r.dim = rend.NewStyle().Foreground(lipgloss.Color("8"))
	r.ok = rend.NewStyle().Foreground(lipgloss.Color("2"))
	r.warn = rend.NewStyle().Foreground(lipgloss.Color("3"))
	r.accent = rend.NewStyle().Foreground(lipgloss.Color("6"))
	r.name = rend.NewStyle().Bold(true)
	r.head = rend.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	return r
}

func (r *reporter) width() int {
	if r.fd != 0 {
		if w, _, err := term.GetSize(r.fd); err == nil && w > 0 {
			return w
		}
	}
	return 100
}

func (r *reporter) header(source string) {
	title := "dl-organizer"
	if r.dryRun {
		title += " (dry-run)"
	}
	fmt.Fprintf(r.out, "%s %s\n\n", r.head.Render(title), r.dim.Render(source+" · "+fmt.Sprint(r.total)+" itens"))
}

// line prints a permanent line above the live bar.
func (r *reporter) line(s string) {
	if r.live {
		fmt.Fprint(r.out, "\r\033[K")
	}
	fmt.Fprintln(r.out, ansi.Truncate(s, r.width()-1, "…"))
	r.bar()
}

func (r *reporter) bar() {
	if !r.live {
		return
	}
	const barWidth = 24
	pct := 100
	filled := barWidth
	if r.total > 0 {
		pct = r.done * 100 / r.total
		filled = r.done * barWidth / r.total
	}
	b := r.accent.Render(strings.Repeat("━", filled)) + r.dim.Render(strings.Repeat("─", barWidth-filled))
	s := fmt.Sprintf("%s %3d%% %s", b, pct, r.dim.Render(fmt.Sprintf("(%d/%d)", r.done, r.total)))
	fmt.Fprint(r.out, "\r\033[K"+ansi.Truncate(s, r.width()-1, ""))
}

func (r *reporter) recordMove(name, target, category string, isDir bool) {
	r.done++
	r.moved++
	r.counts[category]++
	kind := "📄"
	if isDir {
		kind = "📁"
	}
	verb := r.ok.Render("→")
	if r.dryRun {
		verb = r.warn.Render("⇢")
	}
	r.line(fmt.Sprintf(" %s %s %s %s %s", kind, r.name.Render(ansi.Truncate(name, 40, "…")), verb, iconForCategory(category), r.accent.Render(target)))
}

func (r *reporter) skipped(name, reason string) {
	r.done++
	r.skip++
	r.line(r.dim.Render(fmt.Sprintf(" ⏭  %s  (%s)", name, reason)))
}

func (r *reporter) finish() {
	if r.live {
		fmt.Fprint(r.out, "\r\033[K")
	}
	fmt.Fprintln(r.out)
	if len(r.counts) > 0 {
		cats := make([]string, 0, len(r.counts))
		for c := range r.counts {
			cats = append(cats, c)
		}
		sort.Slice(cats, func(i, j int) bool {
			if r.counts[cats[i]] != r.counts[cats[j]] {
				return r.counts[cats[i]] > r.counts[cats[j]]
			}
			return cats[i] < cats[j]
		})
		for _, c := range cats {
			fmt.Fprintf(r.out, " %s %-11s %s\n", iconForCategory(c), c, r.accent.Render(fmt.Sprint(r.counts[c])))
		}
		fmt.Fprintln(r.out)
	}
	label := "movidos"
	if r.dryRun {
		label = "seriam movidos"
	}
	fmt.Fprintf(r.out, "%s %s %s, %d ignorados %s\n", r.ok.Render("✔"), r.name.Render(fmt.Sprint(r.moved)), label, r.skip,
		r.dim.Render(fmt.Sprintf("em %s", time.Since(r.start).Round(time.Millisecond))))
}

// abort ends the live line cleanly before an error is returned.
func (r *reporter) abort() {
	if r.live {
		fmt.Fprint(r.out, "\r\033[K")
	}
}
