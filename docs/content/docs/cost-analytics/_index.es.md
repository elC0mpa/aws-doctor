---
title: "Análisis de Costos"
description: "Descubra los dos flujos de trabajo principales de análisis de costos: Comparativo y de Tendencia."
weight: 30
type: docs
prev: /docs/usage
next: /docs/waste-detection
---

**AWS Doctor** se centra en el análisis de costos contextual. En lugar de mostrar solo números brutos, le ayuda a entender cómo evoluciona su gasto.

## Flujo Comparativo

Para obtener una comparación paralela de su gasto, utilice el subcomando `cost`:

```bash
aws-doctor cost
```

Esto activa el **Flujo Comparativo**, que incluye un desglose por servicio (EC2, S3, etc.) para ayudarle a identificar los impulsores de costos específicos.

### Justicia en la Comparación
Las comparaciones de facturación tradicionales (como mes a mes) suelen ser engañosas. Comparar el 15 de octubre con el mes completo de septiembre siempre parecerá un "ahorro", incluso si está gastando más.

**AWS Doctor** soluciona esto comparando ventanas de tiempo idénticas:
- **Periodo Actual**: 1er día del mes → Hoy.
- **Periodo Anterior**: 1er día del mes anterior → Día idéntico del mes pasado.

*Ejemplo: Si hoy es 15 de octubre, compara del 1 al 15 de octubre con el 1 al 15 de septiembre.*

{{< callout type="warning" >}}
**1er Día del Mes**: Esta función no está disponible el primer día del mes. AWS Cost Explorer requiere un rango mínimo de 24 horas donde la fecha de inicio sea estrictamente anterior a la fecha de finalización.
{{< /callout >}}

---

## Análisis de Tendencias de 6 Meses

Para detectar patrones de crecimiento a largo plazo o cambios arquitectónicos repentinos, utilice el subcomando `trend`:

```bash
aws-doctor trend
```

![Análisis de Tendencias de 6 Meses](/images/demo/trend.gif)

### Qué muestra:
- Un gráfico de barras ANSI de alta fidelidad en su terminal.
- Costos totales mensuales para los últimos 6 ciclos de facturación completados.
- Indicadores claros de si su gasto se está acelerando o estabilizando.
