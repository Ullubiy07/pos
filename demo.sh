#!/usr/bin/env bash
set -euo pipefail

cd -- "$(dirname -- "${BASH_SOURCE[0]}")"

if (( $# > 1 )); then
    printf 'Использование: bash demo.sh [путь-к-новому.csv]\n' >&2
    exit 1
fi

demo_dir=$(mktemp -d "${TMPDIR:-/tmp}/pos-demo.XXXXXX")
csv_path=${1:-"$demo_dir/result.csv"}
log_path="$demo_dir/console.log"

if [[ -e "$csv_path" ]]; then
    printf 'Файл уже существует: %s. Укажите новый путь.\n' "$csv_path" >&2
    exit 1
fi

printf 'Демонстрация: студент → рубрикатор → оценки → итог → CSV\n'
printf 'Ожидаемый итог: 8/10 × 40 + 18/20 × 30 + 9/10 × 30 = 86 / 100\n\n'

# Последовательность ответов для сквозного сценария через меню.
go run ./cmd/pos <<EOF | tee "$log_path"
1
Иван Петров
3
КПО — демонстрационный рубрикатор
3
Домашние задания
10
40
Проект
20
30
Экзамен
10
30
5
6
8
1
8
Домашние задания выполнены
8
2
18
Проект защищён
8
3
9
Экзамен сдан
7
9
7
11
12
$csv_path
0
EOF

log_text=$(<"$log_path")
if [[ "$log_text" == *"Ошибка:"* || "$log_text" != *"Итог: 86 / 100"* || ! -s "$csv_path" ]]; then
    printf '\nДемонстрация не завершилась успешно. Лог: %s\n' "$log_path" >&2
    exit 1
fi

printf '\nДемонстрация завершена. Содержимое CSV:\n'
cat -- "$csv_path"
printf '\nCSV: %s\nЛог консоли: %s\n' "$csv_path" "$log_path"
