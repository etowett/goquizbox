<script>
	export let errors;
</script>

{#if errors}
  <div class="alert alert-danger alert-dismissible fade show" role="alert">
    <button type="button" class="btn-close" data-bs-dismiss="alert" aria-label="Close"></button>
    <ul class="error-messages">
      {#each Object.keys(errors) as key}
        <li>{errors[key]}</li>
      {/each}
    </ul>
  </div>
{/if}
