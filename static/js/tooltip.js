window.tooltip = {
  element: null,

  init() {
    this.element = document.getElementById('tooltip');
  },

  show: (element) => {
    const count = element.getAttribute('data-count');
    const pos = element.getAttribute('data-compromise-pos');
    const categories = pos ? pos.split(',').join(', ') : 'Unknown';

    tooltip.element.innerHTML = `Count: ${count} | ${categories}`;
    tooltip.element.style.display = 'block';

    const rect = element.getBoundingClientRect();
    tooltip.element.style.left = `${rect.left + window.scrollX}px`;
    tooltip.element.style.top = `${rect.top + window.scrollY - 40}px`;
  },

  hide: () => {
    tooltip.element.style.display = 'none';
  }
};