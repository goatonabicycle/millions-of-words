//TODO: Use this on the home page as well.

function updateDisplay(item) {
  const display = document.getElementById('stats-display');
  const customContent = item.dataset.content;

  console.log({ customContent })
  if (customContent) {
    display.innerHTML = customContent;
  } else {
    display.innerHTML = `
          <div class="text-7xl font-bold text-white">${item.dataset.value}</div>
          <div class="text-sm text-gray-400">${item.dataset.desc}</div>
      `;
  }
}

const firstItem = document.querySelector('[data-value]');
updateDisplay(firstItem);

document.querySelectorAll('[data-value]').forEach(item => {
  item.addEventListener('mouseenter', () => {
    document.querySelectorAll('[data-value]').forEach(i => {
      i.classList.remove('bg-gray-700/50', 'text-white');
    });
    item.classList.add('bg-gray-700/50', 'text-white');
    updateDisplay(item);
  });
});