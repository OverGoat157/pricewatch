<script setup>
import { computed } from 'vue'
import { Line } from 'vue-chartjs'
import {
  Chart as ChartJS,
  LineElement,
  PointElement,
  LinearScale,
  CategoryScale,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import { formatRub, formatShortDate } from '../format'

ChartJS.register(LineElement, PointElement, LinearScale, CategoryScale, Tooltip, Legend, Filler)

const props = defineProps({
  history: { type: Array, required: true },
  target: { type: Number, required: true },
})

const chartData = computed(() => ({
  labels: props.history.map((p) => formatShortDate(p.checked_at)),
  datasets: [
    {
      label: 'Цена',
      data: props.history.map((p) => p.price),
      borderColor: '#4f46e5',
      backgroundColor: 'rgba(79, 70, 229, 0.08)',
      fill: true,
      tension: 0.25,
      pointRadius: 2,
      borderWidth: 2,
    },
    {
      label: 'Целевая цена',
      data: props.history.map(() => props.target),
      borderColor: '#16a34a',
      borderDash: [6, 6],
      pointRadius: 0,
      borderWidth: 1.5,
      fill: false,
    },
  ],
}))

const options = {
  responsive: true,
  maintainAspectRatio: false,
  interaction: { mode: 'index', intersect: false },
  plugins: {
    legend: { display: true, labels: { boxWidth: 14 } },
    tooltip: {
      callbacks: { label: (ctx) => `${ctx.dataset.label}: ${formatRub(ctx.parsed.y)}` },
    },
  },
  scales: {
    y: { ticks: { callback: (v) => formatRub(v) } },
  },
}
</script>

<template>
  <div class="chart-wrap">
    <Line :data="chartData" :options="options" />
  </div>
</template>

<style scoped>
.chart-wrap {
  height: 320px;
}
</style>
