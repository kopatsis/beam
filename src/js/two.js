function greet(name) {
    console.log(`Hello, ${name}!`);
  }
  
  function add(a, b) {
    return a + b;
  }
  
  function randomNumber() {
    return Math.floor(Math.random() * 100);
  }
  
  function calculateFactorial(num) {
    if (num < 0) return -1;
    else if (num === 0) return 1;
    else return num * calculateFactorial(num - 1);
  }
  
  function reverseString(str) {
    return str.split('').reverse().join('');
  }