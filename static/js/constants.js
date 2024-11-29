const COLORS = {
  noun: '#ffb3aa',       // nouns, including proper nouns
  verb: '#90caf9',       // all verb forms
  adjective: '#a5d6a7',  // descriptive words
  adverb: '#fff59d',     // modifiers
  pronoun: '#e1bee7',    // personal, possessive, etc.
  preposition: '#ffcc80', // relationship words
  conjunction: '#ff8a65', // connecting words
  determiner: '#80deea',  // articles, numbers, quantifiers
  interjection: '#ffab91' // exclamations
};

const DOM_CLASSES = {
  word: 'word',
  trackInfoItem: 'track-info-item',
  highlightedAll: 'highlighted-all'
};

const DOM_ATTRIBUTES = {
  word: 'data-word',
  track: 'data-track',
  category: 'data-category',
  trackIndex: 'data-track-index',
  count: 'data-count',
  compromisePos: 'data-compromise-pos'
};

const cleanWord = (word) => word.toLowerCase().replace(/^[^a-zA-Z0-9À-ž'-]+|[^a-zA-Z0-9À-ž'-]+$/g, '');
const escapeSelector = (word) => word.replace(/([!"#$%&'()*+,.\/:;<=>?@[\\\]^`{|}~])/g, '\\$1');