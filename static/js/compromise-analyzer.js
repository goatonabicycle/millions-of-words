const POS_DESCRIPTIONS = {
  noun: "Words for people, places, things, or ideas (e.g., 'tree', 'freedom', 'Alex')",
  verb: "Action or state words (e.g., 'run', 'think', 'is')",
  adjective: "Words that describe nouns (e.g., 'blue', 'tall', 'angry')",
  adverb: "Words that modify verbs, adjectives, or other adverbs (e.g., 'quickly', 'very')",
  pronoun: "Words that replace nouns (e.g., 'he', 'which', 'someone')",
  preposition: "Words that show relationships (e.g., 'in', 'on', 'with')",
  conjunction: "Words that connect other words or phrases (e.g., 'and', 'but')",
  determiner: "Words that introduce nouns (e.g., 'the', 'some', 'many')",
  interjection: "Exclamations or expressions (e.g., 'oh!', 'wow')"
};

const compromiseAnalyzer = {
  ...posAnalyzer,

  createPOSContainer(posCategories, trackIndex) {
    const fragment = document.createDocumentFragment();

    const sortedCategories = Object.entries(posCategories)
      .filter(([category, data]) => {
        return category !== '_debug' && data.total > 0;
      })
      .sort((a, b) => b[1].total - a[1].total);

    sortedCategories.forEach(([category, data]) => {
      const div = document.createElement('div');
      div.className = 'pos-tag';
      div.setAttribute(DOM_ATTRIBUTES.category, category);
      div.setAttribute(DOM_ATTRIBUTES.trackIndex, trackIndex);
      div.setAttribute('title', POS_DESCRIPTIONS[category]);

      const baseColor = COLORS[category];
      div.style.background = `linear-gradient(135deg, ${baseColor}40, ${baseColor}50)`;

      div.innerHTML = `
        <span class="pos-category">${category.charAt(0).toUpperCase() + category.slice(1)}</span>
        <span class="pos-count">
          <span class="total-count">${data.total}</span>
        </span>
      `;

      // I'll do something with the unique count later
      // <span class="unique-count font-bold">${data.unique.length}</span>
      // <span class="count-separator text-opacity-75">/</span>

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
      if (!word) return;

      const matchingCategories = [];
      Object.entries(posCategories).forEach(([category, data]) => {
        if (category === '_debug') return;
        if (data.unique && data.unique.includes(word.toLowerCase())) {
          matchingCategories.push(category);
        }
      });

      if (matchingCategories.length > 0) {
        wordElement.setAttribute(DOM_ATTRIBUTES.compromisePos, matchingCategories.join(','));
      }
    });
  }
};

window.compromiseAnalyzer = compromiseAnalyzer;