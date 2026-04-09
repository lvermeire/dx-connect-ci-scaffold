<template>
  <div>
    <form @submit.prevent="addItem">
      <input
        v-model="newName"
        data-testid="item-input"
        placeholder="Item name"
        type="text"
      >
      <button type="submit">
        Add
      </button>
    </form>

    <ul>
      <li
        v-for="item in items"
        :key="item.id"
        data-testid="item"
      >
        {{ item.name }}
      </li>
    </ul>

    <p
      v-if="error"
      data-testid="error"
    >
      {{ error }}
    </p>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const items = ref([])
const newName = ref('')
const error = ref('')

async function fetchItems() {
  try {
    const res = await fetch('/api/items')
    items.value = await res.json()
  } catch {
    error.value = 'Failed to load items'
  }
}

async function addItem() {
  if (!newName.value.trim()) return
  try {
    const res = await fetch('/api/items', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: newName.value.trim() }),
    })
    if (!res.ok) throw new Error('Failed to create item')
    const item = await res.json()
    items.value.push(item)
    newName.value = ''
  } catch (e) {
    error.value = e.message
  }
}

onMounted(fetchItems)
</script>
