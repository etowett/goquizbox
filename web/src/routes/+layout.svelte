<script>
	import { navigating } from '$app/stores';
	import PreloadingIndicator from './PreloadingIndicator.svelte';
	import Nav from './Nav.svelte';
</script>


<div class="container py-3">
  {#if $navigating}
    <PreloadingIndicator />
  {/if}

  <Nav />

  <main class="main-content">
    <slot />
  </main>

  <footer class="pt-4 my-md-5 pt-md-5 border-top">
  <div class="row">
    <div class="col-12 col-md">
      <small class="d-block mb-3 text-muted">GoQuizbox &copy; &middot; 2022</small>
    </div>
  </div>
</footer>
</div>

<style>
  .main-content {
    min-height: 400px;
  }
</style>
