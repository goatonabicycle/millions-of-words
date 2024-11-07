const compromiseAnalyzer = {
  categories: ['noun', 'verb', 'adjective', 'adverb', 'pronoun', 'preposition',
    'conjunction', 'determiner', 'auxiliary', 'particle', 'number',
    'interjection', 'abbreviation'],

  analyze: (text) => {
    const doc = nlp(text.toLowerCase());
    return compromiseAnalyzer.categories.reduce((acc, category) => {
      let words;
      if (category === 'abbreviation') {
        words = doc.acronyms().concat(doc.abbreviations()).out('array');
      } else if (category in doc) {
        words = doc[category + 's']().out('array');
      } else {
        words = doc.match('#' + category.charAt(0).toUpperCase() + category.slice(1)).out('array');
      }
      acc[category] = words;
      return acc;
    }, {});
  },

  createPOSContainer: (posCategories, trackIndex) => {
    const container = document.createElement('div');
    container.className = 'compromise-pos mt-4';

    const content = Object.entries(posCategories)
      .map(([category, words]) => `
              <div class="track-info-item" 
                   data-category="${category}" 
                   data-track-index="${trackIndex}"
                   onmouseover="highlightManager.togglePOSHighlight('${category}', ${trackIndex}, true)"
                   onmouseout="highlightManager.togglePOSHighlight('${category}', ${trackIndex}, false)">
                  <span class="info-title">${category.charAt(0).toUpperCase() + category.slice(1)}s:</span>
                  <span class="info-value">${words.length}</span>
              </div>
          `).join('');

    container.innerHTML = `
          <h4 class="font-semibold m-5">Compromise POS Analysis</h4>
          <div class="track-info-grid second-row">${content}</div>
      `;

    return container;
  },

  attachToTrack: (trackElement, trackIndex, posCategories) => {
    trackElement.querySelectorAll('.word').forEach(wordElement => {
      const word = wordElement.getAttribute('data-word');
      if (word) {
        const category = Object.entries(posCategories)
          .find(([, words]) => words.includes(word.toLowerCase()));
        if (category) {
          wordElement.setAttribute('data-compromise-pos', category[0]);
        }
      }
    });
  }
};

window.compromiseAnalyzer = compromiseAnalyzer;