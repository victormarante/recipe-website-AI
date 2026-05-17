'use strict';

// ── State ────────────────────────────────────────────────────────────────────

let state = {
  recipes: [],
  activeCategory: null,
  searchQuery: '',
  editingId: null,
  pendingDeleteId: null,
};

// ── DOM refs ─────────────────────────────────────────────────────────────────

const $ = (sel, ctx = document) => ctx.querySelector(sel);
const $$ = (sel, ctx = document) => [...ctx.querySelectorAll(sel)];

const categoryList  = $('#category-list');
const recipeCards   = $('#recipe-cards');
const noRecipesMsg  = $('#no-recipes-msg');
const listHeading   = $('#list-heading');
const viewList      = $('#view-list');
const viewDetail    = $('#view-detail');
const recipeDetail  = $('#recipe-detail');
const searchInput   = $('#search-input');
const modalOverlay  = $('#modal-overlay');
const deleteOverlay = $('#delete-overlay');
const recipeForm    = $('#recipe-form');

// ── Icon helper ───────────────────────────────────────────────────────────────

function icon(name) {
  const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
  svg.setAttribute('aria-hidden', 'true');
  svg.classList.add('md-icon');
  const use = document.createElementNS('http://www.w3.org/2000/svg', 'use');
  use.setAttribute('href', `#icon-${name}`);
  svg.appendChild(use);
  return svg;
}

// ── Category helpers ──────────────────────────────────────────────────────────

function getAllCategories(recipes) {
  const set = new Set();
  recipes.forEach(r => (r.categories || []).forEach(c => set.add(c.toLowerCase().trim())));
  return [...set].sort();
}

// ── Rendering ─────────────────────────────────────────────────────────────────

function renderCategories() {
  const categories = getAllCategories(state.recipes);

  categoryList.innerHTML = '';

  const allLi = document.createElement('li');
  const allA  = document.createElement('a');
  allA.href = '#';
  allA.textContent = '← Categories';
  allA.addEventListener('click', e => { e.preventDefault(); showCategoriesHome(); });
  allLi.appendChild(allA);
  categoryList.appendChild(allLi);

  categories.forEach(cat => {
    const li = document.createElement('li');
    const a  = document.createElement('a');
    a.href = '#';
    a.textContent = capitalize(cat);
    a.className = state.activeCategory === cat ? 'active' : '';
    a.addEventListener('click', e => { e.preventDefault(); selectCategory(cat); });
    li.appendChild(a);
    categoryList.appendChild(li);
  });
}

function getFilteredRecipes() {
  let recipes = state.recipes;

  if (state.activeCategory) {
    recipes = recipes.filter(r =>
      (r.categories || []).map(c => c.toLowerCase().trim()).includes(state.activeCategory));
  }

  if (state.searchQuery) {
    const q = state.searchQuery.toLowerCase();
    recipes = recipes.filter(r =>
      r.title.toLowerCase().includes(q) ||
      (r.description || '').toLowerCase().includes(q) ||
      (r.categories || []).some(c => c.toLowerCase().includes(q)) ||
      (r.ingredients || []).some(i => i.toLowerCase().includes(q))
    );
  }

  return recipes;
}

function renderRecipeList() {
  const recipes = getFilteredRecipes();
  recipeCards.innerHTML = '';

  if (recipes.length === 0) {
    noRecipesMsg.classList.remove('hidden');
  } else {
    noRecipesMsg.classList.add('hidden');
    recipes.forEach(r => recipeCards.appendChild(buildCard(r)));
  }

  if (state.searchQuery) {
    listHeading.textContent = `Search: "${state.searchQuery}"`;
  } else if (state.activeCategory) {
    listHeading.textContent = capitalize(state.activeCategory);
  } else {
    listHeading.textContent = 'All Recipes';
  }
}

function buildCard(recipe) {
  const card = document.createElement('div');
  card.className = 'recipe-card';
  card.setAttribute('data-id', recipe.id);

  const h3 = document.createElement('h3');
  h3.textContent = recipe.title;

  const p = document.createElement('p');
  p.textContent = recipe.description || '';

  const tags = document.createElement('div');
  tags.className = 'card-tags';
  (recipe.categories || []).forEach(c => {
    const span = document.createElement('span');
    span.className = 'tag';
    span.textContent = capitalize(c);
    tags.appendChild(span);
  });

  const actions = document.createElement('div');
  actions.className = 'card-actions';

  const editBtn = document.createElement('button');
  editBtn.className = 'btn-icon';
  editBtn.title = 'Edit recipe';
  editBtn.appendChild(icon('edit'));
  editBtn.addEventListener('click', e => { e.stopPropagation(); openEditModal(recipe.id); });

  const delBtn = document.createElement('button');
  delBtn.className = 'btn-icon';
  delBtn.title = 'Delete recipe';
  delBtn.appendChild(icon('delete'));
  delBtn.addEventListener('click', e => { e.stopPropagation(); openDeleteModal(recipe.id); });

  actions.appendChild(editBtn);
  actions.appendChild(delBtn);

  card.appendChild(h3);
  card.appendChild(p);
  card.appendChild(tags);
  card.appendChild(actions);

  card.addEventListener('click', () => showDetail(recipe.id));

  return card;
}

function showDetail(id) {
  const recipe = state.recipes.find(r => r.id === id);
  if (!recipe) return;

  recipeDetail.innerHTML = '';

  const h2 = document.createElement('h2');
  h2.textContent = recipe.title;

  const meta = document.createElement('div');
  meta.className = 'detail-meta';
  (recipe.categories || []).forEach(c => {
    const span = document.createElement('span');
    span.className = 'tag';
    span.textContent = capitalize(c);
    meta.appendChild(span);
  });

  const detailActions = document.createElement('div');
  detailActions.className = 'detail-actions';
  const editBtn = document.createElement('button');
  editBtn.className = 'btn btn-secondary';
  editBtn.appendChild(icon('edit'));
  editBtn.appendChild(document.createTextNode(' Edit'));
  editBtn.addEventListener('click', () => openEditModal(recipe.id));
  const delBtn = document.createElement('button');
  delBtn.className = 'btn btn-danger';
  delBtn.appendChild(icon('delete'));
  delBtn.appendChild(document.createTextNode(' Delete'));
  delBtn.addEventListener('click', () => openDeleteModal(recipe.id));
  detailActions.appendChild(editBtn);
  detailActions.appendChild(delBtn);

  const descSection = document.createElement('section');
  const descP = document.createElement('p');
  descP.textContent = recipe.description || '';
  descSection.appendChild(descP);

  const ingSection = document.createElement('section');
  const ingH3 = document.createElement('h3');
  ingH3.textContent = 'Ingredients';
  const ingUl = document.createElement('ul');
  (recipe.ingredients || []).forEach(ing => {
    const li = document.createElement('li');
    li.textContent = ing;
    ingUl.appendChild(li);
  });
  ingSection.appendChild(ingH3);
  ingSection.appendChild(ingUl);

  const stepsSection = document.createElement('section');
  const stepsH3 = document.createElement('h3');
  stepsH3.textContent = 'Steps';
  const stepsOl = document.createElement('ol');
  (recipe.steps || []).forEach(step => {
    const li = document.createElement('li');
    li.textContent = step;
    stepsOl.appendChild(li);
  });
  stepsSection.appendChild(stepsH3);
  stepsSection.appendChild(stepsOl);

  recipeDetail.appendChild(h2);
  recipeDetail.appendChild(meta);
  recipeDetail.appendChild(detailActions);
  recipeDetail.appendChild(descSection);
  recipeDetail.appendChild(ingSection);
  recipeDetail.appendChild(stepsSection);

  const links = recipe.links || [];
  if (links.length > 0) {
    const linksSection = document.createElement('section');
    const linksH3 = document.createElement('h3');
    linksH3.textContent = 'Related Links';
    const linksDiv = document.createElement('div');
    linksDiv.className = 'detail-links';

    links.forEach(link => {
      const a = document.createElement('a');
      if (link.type === 'recipe') {
        const linked = state.recipes.find(r => r.id === link.linked_recipe_id);
        a.appendChild(icon('link'));
        a.appendChild(document.createTextNode(' ' + (linked ? linked.title : 'Unknown recipe')));
        a.href = '#';
        a.addEventListener('click', e => {
          e.preventDefault();
          if (linked) showDetail(linked.id);
        });
      } else {
        a.appendChild(icon('public'));
        a.appendChild(document.createTextNode(' ' + (link.label || link.url)));
        a.href = link.url;
        a.target = '_blank';
        a.rel = 'noopener noreferrer';
      }
      linksDiv.appendChild(a);
    });

    linksSection.appendChild(linksH3);
    linksSection.appendChild(linksDiv);
    recipeDetail.appendChild(linksSection);
  }

  showView('detail');
}

// ── View switching ────────────────────────────────────────────────────────────

function showView(name) {
  const catHome       = $('#view-categories-home');
  const recipeBrowser = $('#recipe-browser');

  if (name === 'categories') {
    catHome.classList.remove('hidden');
    recipeBrowser.classList.add('hidden');
    return;
  }

  catHome.classList.add('hidden');
  recipeBrowser.classList.remove('hidden');

  viewList.classList.remove('active');
  viewList.classList.add('hidden');
  viewDetail.classList.remove('active');
  viewDetail.classList.add('hidden');

  if (name === 'list') {
    viewList.classList.add('active');
    viewList.classList.remove('hidden');
  } else {
    viewDetail.classList.add('active');
    viewDetail.classList.remove('hidden');
  }
}

function selectCategory(cat) {
  state.activeCategory = cat;
  state.searchQuery = '';
  searchInput.value = '';
  renderCategories();
  renderRecipeList();
  showView('list');
  closeSidebar();
}

function showCategoriesHome() {
  state.activeCategory = null;
  state.searchQuery = '';
  searchInput.value = '';
  renderCategoryCards();
  showView('categories');
}

function renderCategoryCards() {
  const categories = getAllCategories(state.recipes);
  const grid = $('#category-cards-grid');
  const msg  = $('#no-categories-msg');
  grid.innerHTML = '';

  if (categories.length === 0) {
    msg.classList.remove('hidden');
    return;
  }
  msg.classList.add('hidden');

  categories.forEach(cat => {
    const catRecipes = state.recipes.filter(r =>
      (r.categories || []).map(c => c.toLowerCase().trim()).includes(cat));

    const card = document.createElement('div');
    card.className = 'category-home-card';
    card.addEventListener('click', () => selectCategory(cat));

    const emoji = document.createElement('div');
    emoji.className = 'cat-emoji';
    emoji.textContent = getCategoryEmoji(cat);

    const h3 = document.createElement('h3');
    h3.textContent = capitalize(cat);

    const count = document.createElement('p');
    count.className = 'cat-count';
    count.textContent = `${catRecipes.length} recipe${catRecipes.length !== 1 ? 's' : ''}`;

    const preview = document.createElement('ul');
    preview.className = 'cat-preview';
    catRecipes.slice(0, 3).forEach(r => {
      const li = document.createElement('li');
      li.textContent = r.title;
      preview.appendChild(li);
    });

    card.appendChild(emoji);
    card.appendChild(h3);
    card.appendChild(count);
    card.appendChild(preview);
    grid.appendChild(card);
  });
}

function getCategoryEmoji(cat) {
  const map = {
    breakfast: '🍳', lunch: '🥙', dinner: '🍽', vegetarian: '🥦',
    vegan: '🌱', dessert: '🍰', snack: '🍿', soup: '🍲',
    pasta: '🍝', pizza: '🍕', salad: '🥗', meat: '🥩',
    fish: '🐟', seafood: '🦐', baking: '🥖', bread: '🍞',
    drinks: '🥤', cocktail: '🍹'
  };
  return map[cat.toLowerCase()] || '🍴';
}

// ── Modal: Add / Edit ─────────────────────────────────────────────────────────

function openAddModal() {
  state.editingId = null;
  $('#modal-title').textContent = 'Add Recipe';
  recipeForm.reset();
  $('#form-id').value = '';
  clearDynamicList('ingredients-list');
  clearDynamicList('steps-list');
  clearDynamicList('links-list');
  addIngredientRow('');
  addStepRow('');
  modalOverlay.classList.remove('hidden');
  $('#form-title').focus();
}

function openEditModal(id) {
  const recipe = state.recipes.find(r => r.id === id);
  if (!recipe) return;
  state.editingId = id;
  $('#modal-title').textContent = 'Edit Recipe';
  $('#form-id').value = id;
  $('#form-title').value = recipe.title;
  $('#form-description').value = recipe.description || '';
  $('#form-categories').value = (recipe.categories || []).join(', ');
  $('#form-tags').value = '';

  clearDynamicList('ingredients-list');
  (recipe.ingredients || []).forEach(ing => addIngredientRow(ing));
  if ((recipe.ingredients || []).length === 0) addIngredientRow('');

  clearDynamicList('steps-list');
  (recipe.steps || []).forEach(step => addStepRow(step));
  if ((recipe.steps || []).length === 0) addStepRow('');

  clearDynamicList('links-list');
  (recipe.links || []).forEach(link => addLinkRow(link));

  modalOverlay.classList.remove('hidden');
  $('#form-title').focus();
}

function closeModal() {
  modalOverlay.classList.add('hidden');
  clearFormError();
}

function clearDynamicList(listId) {
  document.getElementById(listId).innerHTML = '';
}

function addIngredientRow(value) {
  const list = document.getElementById('ingredients-list');
  const row = buildTextRow(value, 'Ingredient', () => row.remove());
  list.appendChild(row);
}

function addStepRow(value) {
  const list = document.getElementById('steps-list');
  const row = buildTextRow(value, 'Step description', () => row.remove());
  list.appendChild(row);
}

function addLinkRow(link = {}) {
  const list = document.getElementById('links-list');
  const row = document.createElement('div');
  row.className = 'dynamic-item';

  const typeSelect = document.createElement('select');
  typeSelect.className = 'link-type';
  ['external', 'recipe'].forEach(t => {
    const opt = document.createElement('option');
    opt.value = t;
    opt.textContent = t === 'external' ? 'External URL' : 'Recipe';
    typeSelect.appendChild(opt);
  });
  typeSelect.value = link.type || 'external';

  const dlId = 'dl-' + Math.random().toString(36).slice(2);
  const datalist = document.createElement('datalist');
  datalist.id = dlId;

  const urlInput = document.createElement('input');
  urlInput.type = 'text';
  urlInput.className = 'link-url';

  if (link.type === 'recipe') {
    const linked = state.recipes.find(r => r.id === link.linked_recipe_id);
    urlInput.value = linked ? linked.title : '';
    urlInput.placeholder = 'Type recipe name…';
    urlInput.setAttribute('list', dlId);
    state.recipes.forEach(r => {
      const opt = document.createElement('option');
      opt.value = r.title;
      datalist.appendChild(opt);
    });
  } else {
    urlInput.value = link.url || '';
    urlInput.placeholder = 'URL';
  }

  const labelInput = document.createElement('input');
  labelInput.type = 'text';
  labelInput.className = 'link-label';
  labelInput.placeholder = 'Label (optional)';
  labelInput.value = link.label || '';

  typeSelect.addEventListener('change', () => {
    if (typeSelect.value === 'recipe') {
      urlInput.placeholder = 'Type recipe name…';
      urlInput.value = '';
      datalist.innerHTML = '';
      state.recipes.forEach(r => {
        const opt = document.createElement('option');
        opt.value = r.title;
        datalist.appendChild(opt);
      });
      urlInput.setAttribute('list', dlId);
      labelInput.style.display = 'none';
    } else {
      urlInput.placeholder = 'URL';
      urlInput.removeAttribute('list');
      labelInput.style.display = '';
    }
  });
  if (link.type === 'recipe') labelInput.style.display = 'none';

  const removeBtn = document.createElement('button');
  removeBtn.type = 'button';
  removeBtn.className = 'btn-icon';
  removeBtn.title = 'Remove';
  removeBtn.appendChild(icon('close'));
  removeBtn.addEventListener('click', () => row.remove());

  row.appendChild(typeSelect);
  row.appendChild(urlInput);
  row.appendChild(datalist);
  row.appendChild(labelInput);
  row.appendChild(removeBtn);
  list.appendChild(row);
}

function buildTextRow(value, placeholder, onRemove) {
  const row = document.createElement('div');
  row.className = 'dynamic-item';

  const input = document.createElement('input');
  input.type = 'text';
  input.value = value;
  input.placeholder = placeholder;

  const removeBtn = document.createElement('button');
  removeBtn.type = 'button';
  removeBtn.className = 'btn-icon';
  removeBtn.title = 'Remove';
  removeBtn.appendChild(icon('close'));
  removeBtn.addEventListener('click', onRemove);

  row.appendChild(input);
  row.appendChild(removeBtn);
  return row;
}

function collectDynamicValues(listId) {
  return $$(`#${listId} .dynamic-item input[type="text"]`)
    .map(i => i.value.trim())
    .filter(Boolean);
}

function collectLinks() {
  return $$('#links-list .dynamic-item').map(row => {
    const type = row.querySelector('.link-type').value;
    const urlVal = row.querySelector('.link-url').value.trim();
    const labelVal = row.querySelector('.link-label')?.value ?? '';
    if (!urlVal) return null;

    if (type === 'recipe') {
      const match = state.recipes.find(r =>
        r.title.toLowerCase() === urlVal.toLowerCase());
      return match ? { type: 'recipe', linked_recipe_id: match.id } : null;
    }
    return { type: 'external', url: urlVal, label: labelVal.trim() };
  }).filter(Boolean);
}

function showFormError(msg) {
  const el = $('#form-error');
  el.textContent = msg;
  el.classList.remove('hidden');
  el.scrollIntoView({ block: 'nearest' });
}

function clearFormError() {
  const el = $('#form-error');
  el.textContent = '';
  el.classList.add('hidden');
}

async function handleFormSubmit(e) {
  e.preventDefault();
  clearFormError();

  const title = $('#form-title').value.trim();
  if (!title) { showFormError('Please enter a recipe title.'); $('#form-title').focus(); return; }

  const categories = $('#form-categories').value
    .split(',').map(c => c.trim().toLowerCase()).filter(Boolean);
  if (categories.length === 0) { showFormError('Please enter at least one category.'); $('#form-categories').focus(); return; }

  const ingredients = collectDynamicValues('ingredients-list');
  if (ingredients.length === 0) { showFormError('Please add at least one ingredient.'); return; }

  const steps = collectDynamicValues('steps-list');
  if (steps.length === 0) { showFormError('Please add at least one step.'); return; }

  const links = collectLinks();

  const recipeData = {
    title,
    description: $('#form-description').value.trim(),
    categories,
    ingredients,
    steps,
    links,
  };

  const editingId = state.editingId;

  try {
    if (editingId) {
      await window.API.recipes.update(editingId, recipeData);
    } else {
      await window.API.recipes.create(recipeData);
    }
    closeModal();
    await loadAndRender();
    if (editingId && !viewDetail.classList.contains('hidden')) {
      const updated = state.recipes.find(r => r.id === editingId);
      if (updated) showDetail(updated.id);
    }
  } catch (err) {
    showFormError(err.message || 'Failed to save recipe. Please try again.');
  }
}

// ── Modal: Delete ─────────────────────────────────────────────────────────────

function openDeleteModal(id) {
  const recipe = state.recipes.find(r => r.id === id);
  if (!recipe) return;
  state.pendingDeleteId = id;
  $('#delete-recipe-name').textContent = recipe.title;

  const linkedBy = state.recipes.filter(r =>
    r.id !== id && (r.links || []).some(l => l.type === 'recipe' && l.linked_recipe_id === id)
  );
  const blockMsg = $('#delete-block-msg');
  const confirmBtn = $('#btn-confirm-delete');
  if (linkedBy.length > 0) {
    blockMsg.textContent = `Cannot delete: linked by ${linkedBy.map(r => r.title).join(', ')}`;
    blockMsg.classList.remove('hidden');
    confirmBtn.disabled = true;
  } else {
    blockMsg.classList.add('hidden');
    confirmBtn.disabled = false;
  }

  deleteOverlay.classList.remove('hidden');
}

function closeDeleteModal() {
  state.pendingDeleteId = null;
  $('#btn-confirm-delete').disabled = false;
  $('#delete-block-msg').classList.add('hidden');
  deleteOverlay.classList.add('hidden');
}

async function confirmDelete() {
  if (!state.pendingDeleteId) return;

  const linkedBy = state.recipes.filter(r =>
    r.id !== state.pendingDeleteId &&
    (r.links || []).some(l => l.type === 'recipe' && l.linked_recipe_id === state.pendingDeleteId)
  );
  if (linkedBy.length > 0) return;

  try {
    await window.API.recipes.delete(state.pendingDeleteId);
    closeDeleteModal();
    await loadAndRender();
    showView('list');
  } catch (err) {
    const blockMsg = $('#delete-block-msg');
    blockMsg.textContent = err.message || 'Failed to delete recipe.';
    blockMsg.classList.remove('hidden');
  }
}

// ── Utilities ─────────────────────────────────────────────────────────────────

function capitalize(str) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

// ── Mobile sidebar drawer ─────────────────────────────────────────────────────

function toggleSidebar() {
  const sidebar = $('#sidebar');
  const overlay = $('#drawer-overlay');
  const isOpen = sidebar.classList.toggle('open');
  overlay.classList.toggle('hidden', !isOpen);
}

function closeSidebar() {
  $('#sidebar').classList.remove('open');
  $('#drawer-overlay').classList.add('hidden');
}

// ── Event wiring ──────────────────────────────────────────────────────────────

function wireEvents() {
  $('#logo').addEventListener('click', showCategoriesHome);

  $('#btn-logout').addEventListener('click', () => {
    window.API.clearToken();
    location.reload();
  });

  $('#login-form').addEventListener('submit', handleLoginSubmit);

  $('#btn-add-recipe').addEventListener('click', openAddModal);
  $('#btn-cancel').addEventListener('click', closeModal);
  recipeForm.addEventListener('submit', handleFormSubmit);

  $('#btn-add-ingredient').addEventListener('click', () => addIngredientRow(''));
  $('#btn-add-step').addEventListener('click', () => addStepRow(''));
  $('#btn-add-link').addEventListener('click', () => addLinkRow());

  $('#btn-back').addEventListener('click', () => showView('list'));

  $('#btn-confirm-delete').addEventListener('click', confirmDelete);
  $('#btn-cancel-delete').addEventListener('click', closeDeleteModal);

  $('#btn-menu-toggle').addEventListener('click', toggleSidebar);
  $('#drawer-overlay').addEventListener('click', closeSidebar);

  $('#fab-add').addEventListener('click', openAddModal);

  modalOverlay.addEventListener('click', e => { if (e.target === modalOverlay) closeModal(); });
  deleteOverlay.addEventListener('click', e => { if (e.target === deleteOverlay) closeDeleteModal(); });

  document.addEventListener('keydown', e => {
    if (e.key === 'Escape') { closeModal(); closeDeleteModal(); closeSidebar(); }
  });

  searchInput.addEventListener('input', () => {
    state.searchQuery = searchInput.value.trim();
    if (state.searchQuery) {
      state.activeCategory = null;
      renderCategories();
      renderRecipeList();
      showView('list');
    } else {
      showCategoriesHome();
    }
  });
}

// ── Auth ──────────────────────────────────────────────────────────────────────

async function handleLoginSubmit(e) {
  e.preventDefault();
  const username = $('#login-username').value.trim();
  const password = $('#login-password').value;
  const errorEl = $('#login-error');
  errorEl.classList.add('hidden');

  try {
    const { token } = await window.API.auth.login(username, password);
    window.API.setToken(token);
    $('#login-overlay').classList.add('hidden');
    await loadAndRender();
  } catch {
    errorEl.classList.remove('hidden');
  }
}

// ── Data loading ──────────────────────────────────────────────────────────────

async function loadAndRender() {
  try {
    state.recipes = await window.API.recipes.getAll() || [];
  } catch {
    state.recipes = [];
  }
  renderCategories();
  renderCategoryCards();
  renderRecipeList();
  showView('categories');
}

// ── Bootstrap ─────────────────────────────────────────────────────────────────

async function init() {
  wireEvents();
  if (window.API.getToken()) {
    $('#login-overlay').classList.add('hidden');
    await loadAndRender();
  }
}

document.addEventListener('DOMContentLoaded', () => init());
