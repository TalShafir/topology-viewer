package util

import (
	"io"
	"maps"
	"slices"

	"github.com/TalShafir/topology-viewer/pkg/cmd"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/printers"
)

func PrintTopologies(topologies map[string]*cmd.Toplogy, writer io.Writer) {
	colNames := append([]string{"Resource"}, slices.Sorted(maps.Keys(topologies))...)
	cols := make([]metav1.TableColumnDefinition, 0, len(colNames))
	rows := make([]metav1.TableRow, 0)

	for _, k := range colNames {
		cols = append(cols, metav1.TableColumnDefinition{Name: k, Type: "string"})
	}

	rows = append(rows, buildCountRow(topologies, colNames))
	rows = append(rows, buildResourcesRows(topologies, colNames)...)

	table := &metav1.Table{
		ColumnDefinitions: cols,
		Rows:              rows,
	}

	printer := printers.NewTablePrinter(printers.PrintOptions{})
	printer.PrintObj(table, writer)
}

func buildCountRow(topologies map[string]*cmd.Toplogy, colNames []string) metav1.TableRow {
	cells := make([]interface{}, 0, len(colNames))
	cells = append(cells, "count")

	for _, k := range colNames[1:] {
		cells = append(cells, topologies[k].Count)
	}

	row := metav1.TableRow{
		Cells: cells,
	}

	return row
}

func buildResourcesRows(topologies map[string]*cmd.Toplogy, colNames []string) []metav1.TableRow {
	resources := extractResourceNames(topologies)

	rows := make([]metav1.TableRow, 0, len(resources))

	for _, r := range resources {
		cells := make([]interface{}, 0, len(colNames))
		cells = append(cells, r)

		for _, t := range topologies {
			val := "-"
			r, exists := t.Resources[corev1.ResourceName(r)]
			if exists {
				val = r.String()
			}

			cells = append(cells, val)
		}

		row := metav1.TableRow{
			Cells: cells,
		}
		rows = append(rows, row)
	}

	return rows
}

func extractResourceNames(topologies map[string]*cmd.Toplogy) []string {
	rset := sets.New[string]()

	for _, v := range topologies {
		for r := range v.Resources {
			rset.Insert(r.String())
		}
	}

	resources := rset.UnsortedList()
	slices.Sort(resources)

	return resources
}
