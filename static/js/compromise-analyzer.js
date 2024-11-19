const compromiseAnalyzer = {
  get categories() {
    return Object.keys(COLORS);
  },

  analyze(text) {
    const doc = nlp(text.toLowerCase());
    return this.categories.reduce((acc, category) => {
      let words;
      if (category in doc) {
        words = doc[category + 's']().out('array');
      } else {
        words = doc.match('#' + category.charAt(0).toUpperCase() + category.slice(1)).out('array');
      }
      acc[category] = words;
      return acc;
    }, {});
  },

  createPOSContainer(posCategories, trackIndex) {
    const container = document.createElement('div');
    container.className = 'compromise-pos mt-4';

    const content = Object.entries(posCategories)
      .filter(([_, words]) => words.length > 0)
      .map(([category, words]) => `
        <div class="${DOM_CLASSES.trackInfoItem}"
             ${DOM_ATTRIBUTES.category}="${category}"
             ${DOM_ATTRIBUTES.trackIndex}="${trackIndex}"
             onmouseover="highlightManager.togglePOSHighlight('${category}', ${trackIndex}, true)"
             onmouseout="highlightManager.togglePOSHighlight('${category}', ${trackIndex}, false)"
             style="background-color: #374151; color: #d1d5db">
          <span class="info-title">${category.charAt(0).toUpperCase() + category.slice(1)}s:</span>
          <span class="info-value">${words.length}</span>
        </div>
      `).join('');

    container.innerHTML = `
      <h4 class="font-semibold m-5">Parts</h4>
      <div class="track-info-grid second-row">${content}</div>
    `;

    return container;
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