const compromiseAnalyzer = {
  ...posAnalyzer,

  createPOSContainer(posCategories, trackIndex) {
    const fragment = document.createDocumentFragment();

    const sortedCategories = Object.entries(posCategories)
      .filter(([_, words]) => words.length > 0)
      .sort((a, b) => b[1].length - a[1].length);

    sortedCategories.forEach(([category, words]) => {
      const div = document.createElement('div');
      div.className = 'pos-tag';
      div.setAttribute(DOM_ATTRIBUTES.category, category);
      div.setAttribute(DOM_ATTRIBUTES.trackIndex, trackIndex);

      const baseColor = COLORS[category];
      div.style.background = `linear-gradient(135deg, ${baseColor}40, ${baseColor}50)`;

      div.innerHTML = `
              <span class="pos-category">${category.charAt(0).toUpperCase() + category.slice(1)}</span>
              <span class="pos-count">${words.length}</span>
          `;

      div.addEventListener('mouseover', () => {
        highlightManager.togglePOSHighlight(category, trackIndex, true);
      });

      div.addEventListener('mouseout', () => {
        highlightManager.togglePOSHighlight(category, trackIndex, false);
      });

      fragment.appendChild(div);
    });

    return fragment;
  },

  attachToTrack(trackElement, trackIndex, posCategories) {
    trackElement.querySelectorAll(`.${DOM_CLASSES.word}`).forEach(wordElement => {
      const word = wordElement.getAttribute(DOM_ATTRIBUTES.word);
      if (word) {
        const matchingCategories = [];
        Object.entries(posCategories).forEach(([category, words]) => {
          if (words.includes(word.toLowerCase())) {
            matchingCategories.push(category);
          }
        });

        if (matchingCategories.length > 0) {
          wordElement.setAttribute(DOM_ATTRIBUTES.compromisePos, matchingCategories.join(','));
        }
      }
    });
  }
};

window.compromiseAnalyzer = compromiseAnalyzer;