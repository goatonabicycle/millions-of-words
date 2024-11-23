const compromiseAnalyzer = {
  get categories() {
    return Object.keys(COLORS);
  },

  analyze(text) {
    const doc = nlp(text.toLowerCase());
    const debugData = { uncategorized: {} };
    const results = this.categories.reduce((acc, category) => {
      acc[category] = [];
      return acc;
    }, {});

    const allTerms = doc.json();
    allTerms.forEach(sentence => {
      sentence.terms.forEach(term => {
        const word = term.text.toLowerCase();
        const tags = term.tags || [];
        let categorized = false;

        if (word === 'there' || word.startsWith('there')) {
          results.auxiliary.push(word);
          categorized = true;
        } else if (word.includes('-')) {
          if (tags.includes('Noun')) results.noun.push(word);
          else if (tags.includes('Adjective')) results.adjective.push(word);
          else if (tags.includes('Verb')) results.verb.push(word);
          categorized = true;
        } else if (tags.some(tag => ['Value', 'Cardinal', 'TextValue', 'Date', 'Duration'].includes(tag))) {
          results.determiner.push(word);
          categorized = true;
        } else if (tags.includes('Pronoun') || (tags.includes('Possessive') && !tags.includes('Noun'))) {
          results.pronoun.push(word);
          categorized = true;
        } else if (tags.includes('Preposition')) {
          results.preposition.push(word);
          categorized = true;
        } else if (tags.includes('Conjunction')) {
          results.conjunction.push(word);
          categorized = true;
        } else if (tags.includes('Determiner') || word === 'the') {
          results.determiner.push(word);
          categorized = true;
        } else if (tags.includes('Verb') || tags.includes('Infinitive') || tags.includes('PastTense')) {
          results.verb.push(word);
          categorized = true;
        } else if (tags.includes('Adverb') || tags.includes('Negative')) {
          results.adverb.push(word);
          categorized = true;
        } else if (tags.includes('Adjective')) {
          results.adjective.push(word);
          categorized = true;
        } else if (tags.includes('Expression')) {
          results.interjection.push(word);
          categorized = true;
        } else if (tags.includes('Noun') || tags.includes('Singular') || tags.includes('Plural')) {
          results.noun.push(word);
          categorized = true;
        }

        if (!categorized) {
          debugData.uncategorized[word] = tags;
        }
      });
    });

    console.log('Uncategorized words:', debugData.uncategorized);
    return results;
  },

  createPOSContainer(posCategories, trackIndex) {
    const fragment = document.createDocumentFragment();

    Object.entries(posCategories)
      .filter(([_, words]) => words.length > 0)
      .forEach(([category, words]) => {
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