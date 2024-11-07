const cleanWord = (word) => word.toLowerCase().replace(/^[^a-zA-Z0-9À-ž'-]+|[^a-zA-Z0-9À-ž'-]+$/g, '');
const escapeSelector = (word) => word.replace(/([!"#$%&'()*+,.\/:;<=>?@[\\\]^`{|}~])/g, '\\$1');

const highlightManager = {
  posClasses: {
    noun: 'highlighted-noun',
    verb: 'highlighted-verb',
    adjective: 'highlighted-adjective',
    adverb: 'highlighted-adverb',
    pronoun: 'highlighted-pronoun',
    preposition: 'highlighted-preposition',
    conjunction: 'highlighted-conjunction',
    determiner: 'highlighted-determiner',
    auxiliary: 'highlighted-auxiliary',
    particle: 'highlighted-particle',
    number: 'highlighted-number',
    interjection: 'highlighted-interjection',
    abbreviation: 'highlighted-abbreviation'
  },

  toggleHighlight: (elements, className, shouldAdd) => {
    elements.forEach(el => el.classList.toggle(className, shouldAdd));
  },

  toggleWordHighlight: (element, shouldAdd = true) => {
    const { word, track } = element.dataset;
    const selector = `[data-word="${escapeSelector(word)}"][data-track="${track}"]`;
    highlightManager.toggleHighlight(
      document.querySelectorAll(selector),
      'highlighted',
      shouldAdd
    );
  },

  togglePOSHighlight: (category, trackIndex, shouldAdd = true) => {
    const className = highlightManager.posClasses[category] || `highlighted-${category}`;
    const selector = `.word[data-compromise-pos="${category}"][data-track="${trackIndex}"]`;

    highlightManager.toggleHighlight(
      document.querySelectorAll(selector),
      className,
      shouldAdd
    );

    const counter = document.querySelector(
      `.compromise-pos .track-info-item[data-category="${category}"][data-track-index="${trackIndex}"]`
    );
    if (counter) {
      counter.classList.toggle(className, shouldAdd);
    }
  }
};

window.highlightManager = highlightManager;