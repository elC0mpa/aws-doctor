---
title: "AWS Doctor"
layout: "hextra-home"
---

{{< hextra/hero-container
  image="/images/logo.webp"
  imageTitle="AWS Doctor"
  imageWidth="512"
>}}
{{< hextra/hero-badge link="https://github.com/elC0mpa/aws-doctor/releases" >}}
  <div class="hx-w-2 hx-h-2 hx-rounded-full hx-bg-primary-400"></div>
  <span>Última versión: {{< latest-version >}}</span>
  {{< icon name="arrow-circle-right" attributes="height=14" >}}
{{< /hextra/hero-badge >}}

<div class="hx-mt-6 hx-mb-6 hx:mt-6">
{{< hextra/hero-headline >}}
  AWS Doctor
{{< /hextra/hero-headline >}}
</div>

<div class="hx:mt-6 hx-mb-6">
{{< hextra/hero-subtitle >}}
  Potente CLI de código abierto para auditar seguridad, costos y mejores prácticas en AWS.
{{< /hextra/hero-subtitle >}}
</div>

{{< hero-buttons >}}
{{< hextra/hero-button text="Empezar" link="docs/" >}}
{{< hextra/hero-badge style="display: flex; justify-content: center; padding: 13px 12px !important; font-size: .875rem !important;" link="https://github.com/elC0mpa/aws-doctor" >}}
  <span>Ver en GitHub <img class="not-prose" style="display: inline; height: 22px; margin-left: 8px;" src='https://img.shields.io/github/stars/elC0mpa/aws-doctor?style=social'/></span>
  {{< icon name="arrow-circle-right" attributes="height=14" >}}
{{< /hextra/hero-badge >}}
{{< /hero-buttons >}}
{{< /hextra/hero-container >}}

<div class="hx:mt-12"></div>

{{< hextra/hero-section >}}
  Características Principales
{{< /hextra/hero-section >}}

<div class="hx:mt-4"></div>

{{< hextra/feature-grid cols="4" >}}
  {{< hextra/feature-card
    icon="trending-up"
    title="Análisis de Costos"
    subtitle="Obtenga una evaluación justa de su velocidad de gasto. AWS Doctor compara los costos del mes actual con el mismo periodo del mes anterior (ej. del 1 al 10), permitiéndole detectar anomalías y picos en tiempo real."
  >}}

  {{< hextra/feature-card
    icon="trash"
    title="Detección 'Zombie'"
    subtitle="Obtenga un chequeo de salud de alto nivel de toda su cuenta de AWS. La herramienta escanea múltiples servicios simultáneamente para identificar recursos inactivos, desconectados u olvidados."
  >}}

  {{< hextra/feature-card
    icon="terminal"
    title="Formatos de Salida"
    subtitle="Elija el formato que mejor se adapte a su flujo de trabajo. Use tablas enriquecidas en la terminal para auditorías manuales rápidas, o genere una salida JSON estructurada para integrarla en sus pipelines de CI/CD."
  >}}

  {{< hextra/feature-card
    icon="key"
    title="Seguridad e IAM"
    subtitle="Soporte completo para roles protegidos por MFA y auditorías proactivas de credenciales IAM."
  >}}
{{< /hextra/feature-grid >}}

<div class="hx:mt-16"></div>

{{< hextra/hero-section >}}
  Auditoría Instantánea de Infraestructura
{{< /hextra/hero-section >}}

<div class="hx:mt-4"></div>

{{< hextra/feature-grid cols="3" >}}
  {{< hextra/feature-card
    icon="server"
    title="Cómputo y EBS"
    subtitle="Detecta instancias EC2 inactivas, volúmenes EBS sin usar y snapshots huérfanos."
  >}}
  {{< hextra/feature-card
    icon="archive"
    title="Almacenamiento S3"
    subtitle="Audita buckets sin políticas de ciclo de vida y limpia cargas multipartes abandonadas."
  >}}
  {{< hextra/feature-card
    icon="share"
    title="Redes"
    subtitle="Identifica IPs Elásticas sin asociar y Load Balancers sin objetivos saludables."
  >}}
{{< /hextra/feature-grid >}}

<div class="hx:mt-16"></div>

{{< hextra/hero-section >}}
  Únete a la Comunidad
{{< /hextra/hero-section >}}

{{< repo-stats contribLabel="Colaboradores" forksLabel="Forks" >}}

{{< hextra/feature-grid cols="2" >}}
  {{< hextra/feature-card
    icon="terminal"
    title="Reportar Errores"
    subtitle="¿Encontraste un error o tienes una idea para una nueva regla? Ayúdanos a mejorar la herramienta abriendo un issue en GitHub."
    link="https://github.com/elC0mpa/aws-doctor/issues"
  >}}
  {{< hextra/feature-card
    icon="github"
    title="Contribuir Código"
    subtitle="¿Listo para contribuir? Aceptamos PRs para nuevas funciones, correcciones y documentación."
    link="https://github.com/elC0mpa/aws-doctor/pulls"
  >}}
{{< /hextra/feature-grid >}}

<div class="hx:mt-24"></div>
