---
title: "Documentación"
description: "Bienvenido a la documentación oficial de AWS Doctor. Explore nuestras guías de análisis de costos, detección de desperdicios y automatización."
sidebar:
  collapsed: false
---

**AWS Doctor** es una herramienta de chequeo de salud basada en terminal para su infraestructura de AWS. Proporciona contexto inmediato sobre su gasto e identifica recursos "zombie" que le están costando dinero.

## Inicio Rápido

```bash
# Obtener un análisis de costos comparativo (Mes actual vs Mes anterior)
aws-doctor cost

# Escanear recursos inactivos y desperdicios
aws-doctor waste

# Ver una tendencia de gasto de 6 meses
aws-doctor trend
```

## Explorar los Módulos

{{< hextra/feature-grid cols="2" >}}
  {{< hextra/feature-card
    icon="presentation-chart-line"
    title="Guía de Uso"
    link="usage/"
    subtitle="Explicación detallada de los subcomandos de la CLI, flags globales, soporte de MFA y gestión de perfiles."
  >}}
  {{< hextra/feature-card
    icon="calculator"
    title="Análisis de Costos"
    link="cost-analytics/"
    subtitle="Comprensión del motor de facturación comparativa y la visualización de tendencias de 6 meses."
  >}}
  {{< hextra/feature-card
    icon="trash"
    title="Detección de Desperdicios"
    link="waste-detection/"
    subtitle="Inmersión profunda en la detección de recursos zombie: EC2, S3, ELB y CloudWatch."
  >}}
  {{< hextra/feature-card
    icon="cpu-chip"
    title="Automatización"
    link="automation/"
    subtitle="Uso de la salida JSON para integrar AWS Doctor en sus pipelines de CI/CD."
  >}}
{{< /hextra/feature-grid >}}

---

## ¿Necesita Ayuda?
¿Busca algo específico? Utilice la barra de búsqueda en la parte superior de la página para encontrar detalles sobre un servicio o subcomando en particular.
