// Staticlint - инструмент для SAST анализа исходного кода приложения
//
// # Поддерживаемые анализаторы
//
// ## Стандартные анализаторы (golang.org/x/tools/go/analysis/passes)
//
// - atomic: Проверяет некорректное использование sync/atomic функций
// - atomicalign: Проверяет выравнивание 64-битных atomic операций
// - bools: Обнаруживает подозрительные операции с булевыми значениями
// - buildssa: Строит SSA форму для других анализаторов
// - buildtag: Проверяет корректность // +build тегов
// - cgocall: Обнаруживает нарушения правил вызовов CGO
// - composite: Проверяет корректность композитных литералов
// - copylock: Обнаруживает копирование мьютексов
// - ctrlflow: Анализирует поток управления
// - deepequalerrors: Проверяет использование errors с deepcopy
// - errorsas: Проверяет корректность использования errors.As
// - fieldalignment: Предлагает оптимизацию выравнивания полей структур
// - findcall: Демонстрационный анализатор поиска вызовов
// - framepointer: Собирает информацию о указателях фреймов
// - httpresponse: Проверяет обработку HTTP ответов
// - ifaceassert: Обнаруживает бесполезные утверждения типов интерфейсов
// - loopclosure: Обнаруживает некорректные захваты переменных в замыканиях
// - lostcancel: Обнаруживает утечки контекстов
// - nilfunc: Обнаруживает сравнения с nil функцией
// - nilness: Анализирует nil-значения в SSA форме
// - pkgfact: Демонстрационный анализатор фактов пакета
// - printf: Проверяет корректность форматных строк
// - reflectvaluecompare: Обнаруживает сравнение reflect.Value с ==
// - shadow: Обнаруживает затенение переменных
// - shift: Проверяет корректность операций сдвига
// - sigchanyzer: Обнаруживает неправильное использование каналов в signal.Notify
// - sortslice: Проверяет корректность использования sort.Slice
// - stdmethods: Проверяет соответствие стандартным интерфейсам
// - stringintconv: Обнаруживает конвертацию строк в числа без ошибок
// - structtag: Проверяет корректность тегов структур
// - testinggoroutine: Обнаруживает некорректные вызовы в горутинах тестов
// - tests: Проверяет корректность тестов
// - timeformat: Проверяет форматирование времени
// - unmarshal: Проверяет передачу указателей в unmarshal
// - unreachable: Обнаруживает недостижимый код
// - unsafeptr: Проверяет корректность преобразований unsafe.Pointer
// - unusedresult: Обнаруживает неиспользуемые возвращаемые значения
// - unusedwrite: Обнаруживает бесполезные записи в переменные
//
// ## Staticcheck SA анализаторы
//
// - SA1000: Invalid regular expression
// - SA1001: Invalid template
// - SA1002: Invalid time format
// - SA1003: Unsupported argument to functions in encoding/binary
// - SA1004: Suspiciously small untyped constant in time.Sleep
// - SA1005: Invalid first argument to exec.Command
// - SA1006: Printf with dynamic first argument and no further arguments
// - SA1007: Invalid URL in net/url.Parse
// - SA1008: Non-canonical key in http.Header map
// - SA1010: (*regexp.Regexp).FindAll called with n == 0
// - SA1011: Various methods in the strings package expect valid UTF-8
// - SA1012: A nil context.Context is being passed to a function
// - SA1013: io.Seeker.Seek is being called with the whence constant as the first argument
// - SA1014: Non-pointer value passed to Unmarshal or Decode
// - SA1015: Using time.Tick in a way that will leak the underlying ticker
// - SA1016: Trapping a signal with signal.Notify that cannot be trapped
// - SA1017: Channels used with os/signal.Notify should be buffered
// - SA1018: strings.Replace called with n == 0
// - SA1019: Using a deprecated function, variable, constant or field
// - SA1020: Using an invalid host:port pair with net.Listen
// - SA1021: Using bytes.Equal to compare two net.IP
// - SA1022: Using (*time.Timer).Reset with a value we can't check
// - SA1023: Modifying the buffer in an io.Writer implementation
// - SA1024: Cutting a slice out of a string that contains non-ASCII characters
// - SA1025: It is not possible to use (*time.Timer).Reset's return value correctly
// - SA1026: Cannot marshal channels or functions
// - SA1027: Atomic access to 64-bit variable must be 64-bit aligned
// - SA1028: sort.Slice can only be used on slices
// - SA1029: Inappropriate use of sync.WaitGroup
// - SA1030: Invalid argument in call to a strconv function
//
// ## Staticcheck анализатор другого класса (ST)
//
// - ST1005: Incorrectly formatted error string
//
// ## Публичные анализаторы
//
// - ineffassign: Обнаруживает неэффективные присваивания переменным
// - errcheck: Проверяет обработку возвращаемых ошибок
//
// ## Собственный анализатор
//
// - nodirectosexit: Запрещает прямой вызов os.Exit в функции main пакета main
//
// # Пример использования
//
// ## Базовая проверка проекта
//
// ./staticlint ./../...
package main

//В ТЗ сказано использовать стандартные анализаторы passes, поэтому взял их все из документации :)
import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"

	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/kisielk/errcheck/errcheck"
)

func main() {

	mychecks := []*analysis.Analyzer{
		//В ТЗ сказано использовать стандартные анализаторы passes, поэтому взял их все из документации :)
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		// Мой анализатор (os.Exit)
		NoDirectOsExitAnalyzer,
	}

	// Все SA анализаторы из staticcheck
	for _, scAnalyzer := range staticcheck.Analyzers {
		if len(scAnalyzer.Analyzer.Name) >= 2 && scAnalyzer.Analyzer.Name[0:2] == "SA" {
			mychecks = append(mychecks, scAnalyzer.Analyzer)
		}
	}

	// Еще один анализатор из класса ST (staticcheck)
	// ST1005 Incorrectly formatted error string
	for _, scAnalyzer := range stylecheck.Analyzers {
		if scAnalyzer.Analyzer.Name == "ST1005" {
			mychecks = append(mychecks, scAnalyzer.Analyzer)
			break
		}
	}

	// Еще два публичных анализатора
	mychecks = append(mychecks, ineffassign.Analyzer) // Обнаруживает неэффективные присваивания
	mychecks = append(mychecks, errcheck.Analyzer)    // Обнаруживает неиспользуемые параметры

	multichecker.Main(
		mychecks...,
	)

}
