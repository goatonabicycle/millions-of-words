const compromiseAnalyzer = {
  TRADITIONAL_MAPPINGS: {
    INTERROGATIVE_PRONOUNS: ['who', 'which', 'what', "what's", 'whom', 'whose'],
    INTERROGATIVE_ADVERBS: ['when', 'where', 'why', 'how'],
    EXISTENTIAL_ADVERBS: ['there', "there's"]
  },

  get categories() {
    return Object.keys(COLORS);
  },

  analyze(text) {
    const doc = nlp(text.toLowerCase());
    const debugData = {
      allWords: {},
      uncategorized: {},
      specialCases: {
        questionWords: {},
        existential: {},
        contractions: {}
      },
      traditionalMappings: {}
    };

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

        // Store all words and their tags for debugging
        if (!debugData.allWords[word]) {
          debugData.allWords[word] = tags;
        }

        // Handle traditional mappings first
        if (this.TRADITIONAL_MAPPINGS.INTERROGATIVE_PRONOUNS.includes(word)) {
          results.pronoun.push(word);
          debugData.traditionalMappings[word] = 'interrogative_pronoun';
          categorized = true;
        }
        else if (this.TRADITIONAL_MAPPINGS.INTERROGATIVE_ADVERBS.includes(word) ||
          this.TRADITIONAL_MAPPINGS.EXISTENTIAL_ADVERBS.includes(word)) {
          results.adverb.push(word);
          debugData.traditionalMappings[word] = word.includes('there') ? 'existential_adverb' : 'interrogative_adverb';
          categorized = true;
        }

        // Store special cases for debugging
        if (tags.includes('QuestionWord')) {
          debugData.specialCases.questionWords[word] = tags;
        }
        if (tags.includes('There')) {
          debugData.specialCases.existential[word] = tags;
        }
        if (word.includes("'")) {
          debugData.specialCases.contractions[word] = tags;
        }

        // Regular categorization logic
        if (!categorized) {
          if (word.includes('-')) {
            if (tags.includes('Noun')) results.noun.push(word);
            if (tags.includes('Adjective')) results.adjective.push(word);
            if (tags.includes('Verb')) results.verb.push(word);
            categorized = true;
          }

          if (tags.some(tag => ['Value', 'Cardinal', 'TextValue', 'Date', 'Duration'].includes(tag))) {
            results.determiner.push(word);
            categorized = true;
          }

          if (tags.includes('Pronoun') || (tags.includes('Possessive') && !tags.includes('Noun'))) {
            results.pronoun.push(word);
            categorized = true;
          }

          if (tags.includes('Preposition')) {
            results.preposition.push(word);
            categorized = true;
          }

          if (tags.includes('Conjunction')) {
            results.conjunction.push(word);
            categorized = true;
          }

          if (tags.includes('Determiner') || word === 'the' || word === 'a' || word === 'an') {
            results.determiner.push(word);
            categorized = true;
          }

          if (tags.includes('Verb') || tags.includes('Infinitive') || tags.includes('PastTense') || tags.includes('Copula')) {
            results.verb.push(word);
            categorized = true;
          }

          if (tags.includes('Adverb') || tags.includes('Negative')) {
            results.adverb.push(word);
            categorized = true;
          }

          if (tags.includes('Adjective')) {
            results.adjective.push(word);
            categorized = true;
          }

          if (tags.includes('Expression') || tags.includes('Interjection')) {
            results.interjection.push(word);
            categorized = true;
          }

          if (tags.includes('Noun') || tags.includes('Singular') || tags.includes('Plural')) {
            results.noun.push(word);
            categorized = true;
          }
        }

        // Track uncategorized words
        if (!categorized) {
          debugData.uncategorized[word] = tags;
        }
      });
    });

    // One-time console output
    if (!window.hasLoggedCompromise) {
      console.log('Compromise Analysis Debug:', JSON.stringify({
        traditionalMappings: debugData.traditionalMappings,
        specialCases: debugData.specialCases,
        uncategorized: debugData.uncategorized,
        allWords: debugData.allWords
      }, null, 2));
      window.hasLoggedCompromise = true;
    }

    return results;
  },

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