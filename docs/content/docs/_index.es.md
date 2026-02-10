---
title: "Documentación"
weight: 1
---

Bienvenido a la documentación de **AWS Doctor**. Esta guía le ayudará a configurar, utilizar y dominar la herramienta para mantener su infraestructura de AWS eficiente, segura y rentable.

## Navegación

<div class="hx:mb-6"></div>

{{< hextra/feature-grid cols="3" >}}
  {{< hextra/feature-card
    icon="terminal"
    title="Primeros Pasos"
    link="getting-started/"
    subtitle="Aprenda cómo instalar AWS Doctor y configurar los permisos mínimos requeridos."
  >}}
  {{< hextra/feature-card
    icon="key"
    title="Guía de Uso"
    link="usage/"
    subtitle="Explicación detallada de los flags de la CLI, soporte de MFA y gestión de perfiles."
  >}}
  {{< hextra/feature-card
    icon="trending-up"
    title="Analítica"
    link="cost-analytics/"
    subtitle="Entienda cómo AWS Doctor realiza comparaciones de costos justas y análisis de tendencias."
  >}}
  {{< hextra/feature-card
    icon="search"
    title="Desperdicio"
    link="waste-detection/"
    subtitle="Lógica técnica detallada para detectar desperdicio en EC2, S3 y Redes."
  >}}
  {{< hextra/feature-card
    icon="server"
    title="Automatización"
    link="automation/"
    subtitle="Guía para la salida JSON e integración con GitHub Actions o Jenkins."
  >}}
{{< /hextra/feature-grid >}}

## Contexto Rápido

- **Sin Estado**: La herramienta nunca almacena sus datos ni credenciales.
- **Evaluación Justa**: Las comparaciones de costos utilizan ventanas de tiempo idénticas para mayor precisión.
- **Configuración Cero**: Funciona directamente con su configuración existente en `~/.aws/config`.

{{< callout type="info" >}}
¿Busca algo específico? Utilice la barra de búsqueda en la parte superior de la página para encontrar detalles sobre un servicio o flag en particular.
{{< /callout >}}
