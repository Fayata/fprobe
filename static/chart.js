// Chart initialization untuk Dashboard
function initChart(historyData) {
    if (!historyData || historyData.length === 0) {
        return;
    }

    // Sort ascending by time for smooth left->right
    const sorted = [...historyData].sort((a, b) => new Date(a.Timestamp) - new Date(b.Timestamp));
    const labels = sorted.map(d => new Date(d.Timestamp).toLocaleString('id-ID', {
        hour: '2-digit', minute: '2-digit'
    }));
    const latencyValues = sorted.map(d => d.LatencyMs);

    const config = {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'Latency (ms)',
                data: latencyValues,
                fill: true,
                borderColor: '#25c17e',
                borderWidth: 2,
                tension: 0.35,
                pointRadius: 0,
                pointHoverRadius: 0,
                segment: {
                    borderColor: ctx => {
                        const i = ctx.p0DataIndex;
                        if (i <= 0) return '#25c17e';
                        const prev = latencyValues[i - 1];
                        const curr = latencyValues[i];
                        return curr >= prev ? '#25c17e' : '#ff6b6b';
                    }
                },
                backgroundColor: (ctx) => {
                    const {ctx: chartCtx, chartArea} = ctx.chart;
                    if (!chartArea) return 'rgba(37, 193, 126, 0.15)';
                    const gradient = chartCtx.createLinearGradient(0, chartArea.top, 0, chartArea.bottom);
                    gradient.addColorStop(0, 'rgba(37, 193, 126, 0.25)');
                    gradient.addColorStop(1, 'rgba(37, 193, 126, 0.00)');
                    return gradient;
                }
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: { display: false },
                tooltip: {
                    backgroundColor: 'rgba(18, 20, 23, 0.95)',
                    titleColor: '#fff',
                    bodyColor: '#fff',
                    borderColor: 'rgba(255,255,255,0.08)',
                    borderWidth: 1,
                    padding: 10,
                    displayColors: false,
                    callbacks: {
                        title: function(items){
                            return 'Latency';
                        },
                        label: function(context) {
                            const value = context.parsed.y;
                            return value + ' ms';
                        }
                    }
                }
            },
            scales: {
                y: {
                    beginAtZero: false,
                    ticks: {
                        color: 'rgba(255, 255, 255, 0.7)',
                        callback: function(value) {
                            return value + ' ms';
                        }
                    },
                    grid: {
                        color: 'rgba(255, 255, 255, 0.06)'
                    }
                },
                x: {
                    ticks: {
                        color: 'rgba(255, 255, 255, 0.6)',
                        maxRotation: 0,
                        minRotation: 0,
                        autoSkip: true,
                        maxTicksLimit: 8
                    },
                    grid: {
                        display: false
                    }
                }
            }
        }
    };

    const ctx = document.getElementById('latencyChart');
    if (ctx) {
        new Chart(ctx.getContext('2d'), config);
    }
}