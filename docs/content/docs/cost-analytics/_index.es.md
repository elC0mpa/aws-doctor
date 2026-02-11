---
title: "Análisis de Costos"
weight: 30
type: docs
prev: /docs/usage
next: /docs/waste-detection
---

**AWS Doctor** proporciona un análisis de costos contextual que va más allá de los simples totales.

{{< callout type="info" >}}
**Permisos Requeridos**: `ce:GetCostAndUsage`
{{< /callout >}}

## Análisis Comparativo de Costos

Cuando ejecuta `aws-doctor` sin flags, se activa el **Flujo Comparativo**. Esto incluye un desglose por servicio (EC2, S3, etc.) para ayudarle a identificar los impulsores de costos específicos.

![Análisis Comparativo de Costos](/images/demo/basic.gif)

### La Lógica de "Evaluación Justa"
La mayoría de las herramientas de facturación comparan el total del mes actual con el total del mes anterior. Esto suele ser engañoso (por ejemplo, comparar 10 días de gasto en febrero con 31 días en enero).

**AWS Doctor** compara ventanas de tiempo idénticas:
- **Periodo Actual**: 1er día del mes actual → Hoy.
- **Periodo Anterior**: 1er día del mes anterior → Día idéntico del mes pasado.

*Ejemplo: Si hoy es 15 de octubre, compara del 1 al 15 de octubre con el 1 al 15 de septiembre.*

{{< callout type="warning" >}}
**1er Día del Mes**: Esta función no está disponible el primer día del mes. AWS Cost Explorer requiere un rango mínimo de 24 horas donde la fecha de inicio sea estrictamente anterior a la fecha de finalización.
{{< /callout >}}

---

## Análisis de Tendencias de 6 Meses

Para detectar patrones de crecimiento a largo plazo o cambios arquitectónicos repentinos, utilice el flag `--trend`:

```bash
aws-doctor --trend
```

![Análisis de Tendencias de 6 Meses](/images/demo/trend.gif)

### Qué muestra:
- Un gráfico de barras ANSI de alta fidelidad en su terminal.
- Costos totales mensuales para los últimos 6 ciclos de facturación completados.
- Indicadores claros de si su gasto se está acelerando o estabilizando.
