// chart.js - Chart initialization untuk Dashboard
function initChart(historyData) {
    if (!historyData || historyData.length === 0) {
        return;
    }

    const labels = historyData.map(d => {
        return new Date(d.Timestamp).toLocaleTimeString('id-ID', { 
            hour: '2-digit', 
            minute: '2-digit', 
            second: '2-digit' 
        });
    }).reverse();

    const latencyValues = historyData.map(d => d.LatencyMs).reverse();

    const config = {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'Latency (ms)',
                data: latencyValues,
                fill: true,
                backgroundColor: 'rgba(198, 40, 40, 0.2)',
                borderColor: '#c62828',
                tension: 0.4,
                pointBackgroundColor: '#c62828',
                pointBorderColor: '#fff',
                pointBorderWidth: 2,
                pointRadius: 4,
                pointHoverRadius: 6
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: true,
            plugins: {
                legend: {
                    display: true,
                    position: 'top',
                    labels: {
                        color: '#fff',
                        font: {
                            size: 14,
                            weight: 'bold'
                        }
                    }
                },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.8)',
                    titleColor: '#fff',
                    bodyColor: '#fff',
                    borderColor: '#c62828',
                    borderWidth: 1,
                    padding: 12,
                    displayColors: false,
                    callbacks: {
                        label: function(context) {
                            return 'Latency: ' + context.parsed.y + ' ms';
                        }
                    }
                }
            },
            scales: {
                y: {
                    beginAtZero: true,
                    ticks: {
                        color: 'rgba(255, 255, 255, 0.7)',
                        callback: function(value) {
                            return value + ' ms';
                        }
                    },
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    }
                },
                x: {
                    ticks: {
                        color: 'rgba(255, 255, 255, 0.7)',
                        maxRotation: 45,
                        minRotation: 45
                    },
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
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