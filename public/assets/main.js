if (!WebAssembly.instantiateStreaming) {
  WebAssembly.instantiateStreaming = async (resp, importObject) => {
    const source = await (await resp).arrayBuffer();
    return await WebAssembly.instantiate(source, importObject);
  };
}

const go = new Go();
WebAssembly.instantiateStreaming(fetch('./assets/optim.wasm'), go.importObject).then(async (result) => {
  go.run(result.instance);
  await main();
});

const main = async () => {
  const form = document.querySelector('form#form');
  const answer = document.querySelector('div#answer');

  const toIcons = (alloc) => {
    const list = [];
    for (let i = 0; i < alloc.s; i++) list.push(`<span class="icon icon-shell"></span>`);
    for (let i = 0; i < alloc.t; i++) list.push(`<span class="icon icon-tool"></span>`);
    for (let i = 0; i < alloc.d; i++) list.push(`<span class="icon icon-demon"></span>`);
    for (let i = 0; i < alloc.c; i++) list.push(`<span class="icon icon-crystal"></span>`);
    return list.join('');
  };

  form.addEventListener('submit', async (e) => {
    e.preventDefault();

    // Clear existing answer
    answer.innerHTML = '';

    const formData = {};
    for (const pair of new FormData(e.target)) {
      const val = pair[1];
      if (val) {
        formData[pair[0]] = parseInt(val);
      }
    }

    if (Object.keys(formData).length === 0) return;

    try {
      const exchangeRates = JSON.parse(await rateTable());

      const bestTrades = await bestTrade(JSON.stringify(formData));
      const optimalTrades = JSON.parse(bestTrades);
      if (optimalTrades.length < 1) {
        throw new Error('No possible trades!');
      }
      const optimalTrade = optimalTrades[0];

      // Sort optimal resources by their trade amount.
      const optimalTradeResources = optimalTrade.R.sort((a, b) => {
        const rateA = exchangeRates[JSON.stringify(a)];
        const rateB = exchangeRates[JSON.stringify(b)];

        if (rateA === rateB) return b.length - a.length;

        return rateB - rateA;
      });

      answer.innerHTML = `
        <hr>

        <table>
          <tr>
            <th align="right" width="100%">Total</th>
            <th>
              <span class="icon icon-coin">${optimalTrade.V}</span>
            </th>
          </tr>
          ${optimalTradeResources
            .map((alloc) => {
              return `
              <tr>
                <td align="right" width="100%">${toIcons(alloc)}</td>
                <td>
                  <span class="icon icon-coin">${exchangeRates[JSON.stringify(alloc)]}</span>
                </td>
              </tr>
            `;
            })
            .join('')}
        </table>
      `;
    } catch (e) {
      console.error(e);
    }
  });
};
