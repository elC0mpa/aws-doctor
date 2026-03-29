---
title: "Reportes PDF"
description: "Genera reportes PDF profesionales para las partes interesadas con un comando. Exporta hallazgos de costos, desperdicio y tendencias en un documento limpio."
weight: 50
type: docs
prev: /docs/waste-detection
next: /docs/automation
---

**AWS Doctor** ahora soporta reportes PDF profesionales. Esta característica le permite transformar hallazgos técnicos en documentos claros y listos para compartir con las partes interesadas y la gerencia.

## Cómo Ejecutar

Use el subcomando `report` seguido del tipo de análisis que desea exportar.

### 1. Reporte de Comparación de Costos
Genere un reporte PDF comparando los gastos del mes actual con el mes anterior.

```bash
aws-doctor report cost
```
{{< pdf "/reports/cost.pdf" >}}

### 2. Reporte de Detección de Desperdicio
Exporte todos los recursos "zombie" identificados (EC2, S3, RDS, etc.) en un documento PDF categorizado.

```bash
aws-doctor report waste
```
{{< pdf "/reports/waste.pdf" >}}

### 3. Reporte de Tendencia de Costos
Genere un reporte visual de tendencia de costos de 6 meses con gráficos de barras e indicadores de aceleración.

```bash
aws-doctor report trend
```
{{< pdf "/reports/trend.pdf" >}}

---

## Flags Globales para Reportes

El subcomando `report` soporta los flags globales estándar (`--profile`, `--region`) y un flag específico para la ubicación de salida.

| Flag | Por Defecto | Descripción |
| :--- | :--- | :--- |
| `--path` | `Documentos` | Especifica el directorio de salida o la ruta completa del archivo para el PDF. |

### Ruta de Salida Personalizada
Por defecto, los reportes se guardan en su carpeta de **Documentos**. En caso de que la carpeta Documentos no exista (común en algunas distribuciones de Linux), se guardará en su carpeta **Home**.

Los archivos se nombran siguiendo el patrón `aws-doctor-<tipo>-<timestamp>.pdf`. Puede sobrescribir esto usando el flag `--path`:

```bash
# Guardar en un directorio específico
aws-doctor report cost --path ./mis-auditorias/

# Guardar con un nombre de archivo específico
aws-doctor report waste --path ./escaneo-semanal.pdf
```

---

## Características del Reporte PDF

- **Diseño Profesional**: Diseño limpio y minimalista optimizado para la lectura.
- **Personalizable**: Utiliza el logo de AWS Doctor y tipografía profesional.
- **Resúmenes Accionables**: Cada sección incluye totales claros y ahorros identificados.
- **Gráficos y Diagramas**: Representaciones visuales de tendencias de costos y distribución de recursos.
- **Hallazgos Categorizados**: Los recursos se agrupan por servicio (Cómputo, Almacenamiento, Redes, etc.) al igual que en la salida de la terminal.

{{< callout type="info" >}}
**AWS Doctor** genera el PDF completamente de forma local. Los datos de su infraestructura nunca salen de su máquina durante el proceso de generación del reporte.
{{< /callout >}}
