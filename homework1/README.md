# Домашнее задание 1: Построение Control Flow Graph (CFG)

## Цель
Научиться извлекать промежуточное представление (IR) из Go компилятора и строить граф потока управления для анализа Go программ.

## ⚠️ Важно
Весь код должен быть реализован в общей кодовой базе: `../internal/cfg/`
Это задание развивает модуль CFG анализа, который будет использоваться в других частях курса.

## Теоретические основы

### Control Flow Graph (CFG)
CFG - это граф, где:
- **Узлы (nodes)** - базовые блоки (sequences of instructions without jumps)
- **Рёбра (edges)** - возможные переходы управления между блоками
- **Entry node** - точка входа в функцию
- **Exit nodes** - точки выхода из функции

### Go SSA IR
Go компилятор предоставляет SSA (Static Single Assignment) представление через пакет `golang.org/x/tools/go/ssa`.

## Задания

### Задание 1.1: Изучение Go SSA
Создайте программу для извлечения SSA представления:

```go
// Пример структуры для анализа
func analyzeFunction(source string, funcName string) (*ssa.Function, error) {
    // 1. Парсинг исходного кода
    // 2. Создание SSA представления  
    // 3. Поиск функции по имени
    // 4. Возврат SSA функции
}
```

**Требования:**
- Используйте пакет `golang.org/x/tools/go/ssa`
- Научитесь получать список инструкций для каждого базового блока
- Выведите информацию о блоках и их связях

### Задание 1.2: Построение CFG для простых функций
Реализуйте построение CFG для функций с условными операторами:

```go
// Тестовые функции для анализа:

func simpleIf(x int) int {
    if x > 0 {
        return x * 2
    }
    return 0
}

func ifElse(x int) int {
    if x > 0 {
        return x * 2
    } else {
        return x * -1
    }
    return 0
}

func nestedIf(x, y int) int {
    if x > 0 {
        if y > 0 {
            return x + y
        }
        return x
    }
    return 0
}
```

**Структура CFG:**
```go
type BasicBlock struct {
    ID           int
    Instructions []ssa.Instruction
    Successors   []*BasicBlock
    Predecessors []*BasicBlock
}

type CFG struct {
    Entry  *BasicBlock
    Blocks []*BasicBlock
    Exit   *BasicBlock
}
```

### Задание 1.3: Анализ циклов в CFG
Добавьте поддержку циклов и постройте CFG для функций с циклами:

```go
func simpleLoop(n int) int {
    sum := 0
    for i := 0; i < n; i++ {
        sum += i
    }
    return sum
}

func whileLoop(x int) int {
    for x > 0 {
        x = x - 1
    }
    return x
}

func nestedLoops(n, m int) int {
    sum := 0
    for i := 0; i < n; i++ {
        for j := 0; j < m; j++ {
            sum += i * j
        }
    }
    return sum
}
```

**Дополнительные задачи:**
- Идентифицируйте back edges (рёбра, создающие циклы)
- Найдите естественные циклы в графе
- Определите loop headers и exit blocks

### Задание 1.4: Визуализация CFG *(необязательно)*
🎯 **Это задание необязательно для выполнения** - реализуйте только если хотите получить дополнительные баллы или вам интересно!

Создайте функцию для экспорта CFG в формат DOT (Graphviz):

```go
func (cfg *CFG) ExportToDot() string {
    // Генерация DOT представления для визуализации
}
```

**Требования:**
- Каждый базовый блок должен показывать свои инструкции
- Рёбра должны быть подписаны (true/false для условных переходов)
- Back edges должны быть выделены другим цветом

## Что сдавать

**Реализация в общей кодовой базе:**
- `../internal/cfg/types.go` - дополните методы для BasicBlock и CFG
- `../internal/cfg/builder.go` - реализуйте парсинг SSA и построение CFG  
- `../internal/cfg/visualizer.go` - *(необязательно)* добавьте экспорт в DOT и статистику

**В папке homework1:**
- `main.go` - демонстрационная программа (уже создана)
- `examples/test_functions.go` - тестовые функции (уже созданы)

```
homework1/
├── examples/
│   └── test_functions.go    // Функции для тестирования CFG
├── main.go                  // Демонстрация (использует internal/cfg)
├── REPORT.md                // Ваш отчёт
└── README.md
```

## Полезные ресурсы
- [golang.org/x/tools/go/ssa](https://pkg.go.dev/golang.org/x/tools/go/ssa)
- [SSA форма в компиляторах](https://en.wikipedia.org/wiki/Static_single_assignment_form)
- [Control Flow Analysis](https://en.wikipedia.org/wiki/Control-flow_analysis)