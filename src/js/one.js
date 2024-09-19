function isPrime(num) {
    if (num <= 1) return false;
    if (num === 2) return true;
    for (let i = 2; i < num; i++) {
      if (num % i === 0) return false;
    }
    return true;
  }
  
  function fibonacci(n) {
    if (n <= 1) return n;
    return fibonacci(n - 1) + fibonacci(n - 2);
  }
  
  function findMax(arr) {
    return Math.max(...arr);
  }
  
  function formatCurrency(amount) {
    return `$${amount.toFixed(2)}`;
  }