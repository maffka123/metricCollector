# Metrics collector service

The service collects a number of metrics from `runtime` and sends them furter to `http://<HOST>/update/<METRIC_TYPE>/<METRIC_NAME>/<VALUE>` and `"application-type": "text/plain"`

[Metrics collector service](#metrics-collector-service)
  * [Metrics types](#metrics-types)
  * [Metrics list](#metrics-list)
  * [Global variables](#global-variables)
  * [Service quit signals](#service-quit-signals)

## Metrics types

* gauge, float64
* counter, int64

Metrics sourse: `runtime`

## Metrics list2

* "Alloc", gauge
* "BuckHashSys", gauge
* "Frees", gauge
* "GCCPUFraction", gauge
* "GCSys", gauge
* "HeapAlloc", gauge
* "HeapIdle", gauge
* "HeapInuse", gauge
* "HeapObjects", gauge
* "HeapReleased", gauge
* "HeapSys", gauge
* "LastGC", gauge
* "Lookups", gauge
* "MCacheInuse", gauge
* "MCacheSys", gauge
* "MSpanInuse", gauge
* "MSpanSys", gauge
* "Mallocs", gauge
* "NextGC", gauge
* "NumForcedGC", gauge
* "NumGC", gauge
* "OtherSys", gauge
* "PauseTotalNs", gauge
* "StackInuse", gauge
* "StackSys", gauge
* "Sys", gauge
* "PollCount", counter — conter incremted by 1, every time metrics are updated from runtime (every pollInterval — see below).
* "RandomValue", gauge — updated random value.

## Global variables

* pollInterval - how oftem metrics are updated (every 2s)
* reportInterval - how often metrics are sent (every 10s)

## Service quit signals

* `syscall.SIGTERM`
* `syscall.SIGINT`
* `syscall.SIGQUIT`

# Обновление шаблона

Чтобы получать обновления автотестов и других частей шаблона, выполните следующую команду:

```
git remote add -m main template https://github.com/yandex-praktikum/go-musthave-devops-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.
