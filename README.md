# go-musthave-metrics-tpl

Шаблон репозитория для трека «Сервер сбора метрик и алертинга».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-metrics-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Профилирование памяти

```bash
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
```

```text
File: result-server
Type: alloc_space
Time: 2026-09-07 16:43:20 MSK
Showing nodes accounting for -16141.54MB, 95.63% of 16879.01MB total
Dropped 201 nodes (cum <= 84.40MB)
      flat  flat%   sum%        cum   cum%
-12484.60MB 73.97% 73.97% -15205.26MB 90.08%  compress/flate.NewWriter (inline)
-2652.12MB 15.71% 89.68% -2652.12MB 15.71%  compress/flate.(*compressor).initDeflate (inline)
 -589.03MB  3.49% 93.17%  -589.03MB  3.49%  compress/flate.(*dictDecoder).init (inline)
 -155.60MB  0.92% 94.09%  -744.62MB  4.41%  compress/flate.NewReader
 -146.32MB  0.87% 94.96%  -146.32MB  0.87%  compress/flate.(*huffmanEncoder).generate
  -96.87MB  0.57% 95.53%   -96.87MB  0.57%  bufio.NewReaderSize (inline)
  -17.01MB   0.1% 95.63%  -842.61MB  4.99%  compress/gzip.NewReader (inline)
    2.50MB 0.015% 95.62% -16241.57MB 96.22%  github.com/shigabutdinoff/metrics/internal/server.(*Server).setupRoutes.WithLogging.func5.1
      -2MB 0.012% 95.63% -15224.80MB 90.20%  github.com/shigabutdinoff/metrics/internal/server.(*Server).setupRoutes.func3.Middleware.2.1
    0.50MB 0.003% 95.62%   103.64MB  0.61%  github.com/shigabutdinoff/metrics/internal/handlers/middleware/reqbody.Read
   -0.50MB 0.003% 95.63% -16299.25MB 96.57%  net/http.(*conn).serve
   -0.50MB 0.003% 95.63%  -843.11MB  5.00%  github.com/shigabutdinoff/metrics/internal/handlers/middleware/compress.newCompressReader
         0     0% 95.63%   -96.87MB  0.57%  bufio.NewReader (inline)
         0     0% 95.63%  -146.32MB  0.87%  compress/flate.(*Writer).Close (inline)
         0     0% 95.63%  -146.32MB  0.87%  compress/flate.(*compressor).close
         0     0% 95.63%  -146.32MB  0.87%  compress/flate.(*compressor).deflate
         0     0% 95.63% -2720.66MB 16.12%  compress/flate.(*compressor).init
         0     0% 95.63%  -146.32MB  0.87%  compress/flate.(*compressor).writeBlock
         0     0% 95.63%   -89.20MB  0.53%  compress/flate.(*huffmanBitWriter).indexTokens
         0     0% 95.63%  -146.32MB  0.87%  compress/flate.(*huffmanBitWriter).writeBlock
         0     0% 95.63%  -819.41MB  4.85%  compress/gzip.(*Reader).Reset
         0     0% 95.63%  -744.62MB  4.41%  compress/gzip.(*Reader).readHeader
         0     0% 95.63%  -146.32MB  0.87%  compress/gzip.(*Writer).Close
         0     0% 95.63% -15205.26MB 90.08%  compress/gzip.(*Writer).Write
         0     0% 95.63% -16212.04MB 96.05%  github.com/go-chi/chi/v5.(*ChainHandler).ServeHTTP
         0     0% 95.63% -16242.57MB 96.23%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 95.63% -16212.04MB 96.05%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 95.63% -16212.04MB 96.05%  github.com/go-chi/chi/v5/middleware.AllowContentType.func1.1
         0     0% 95.63%  -146.32MB  0.87%  github.com/shigabutdinoff/metrics/internal/handlers/middleware/compress.(*compressWriter).Close
         0     0% 95.63% -15205.26MB 90.08%  github.com/shigabutdinoff/metrics/internal/handlers/middleware/compress.(*compressWriter).Write
         0     0% 95.63% -15206.26MB 90.09%  github.com/shigabutdinoff/metrics/internal/handlers/middleware/hash.(*responseWriter).flush
         0     0% 95.63% -16212.04MB 96.05%  github.com/shigabutdinoff/metrics/internal/server.(*Server).setupRoutes.func1.1
         0     0% 95.63% -16212.04MB 96.05%  github.com/shigabutdinoff/metrics/internal/server.(*Server).setupRoutes.func3.GzipMiddleware.1.1
         0     0% 95.63% -16241.57MB 96.22%  net/http.HandlerFunc.ServeHTTP
         0     0% 95.63% -16242.57MB 96.23%  net/http.serverHandler.ServeHTTP
```
