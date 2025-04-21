import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// Конфигурация теста
export const options = {
  stages: [
    { duration: '30s', target: 1000 },  // плавный набор нагрузки
    { duration: '5m', target: 1000 },   // стабильная нагрузка
    { duration: '30s', target: 0 },     // плавное снижение
  ],
  thresholds: {
    http_req_duration: ['p(95)<100'],  // 95% запросов <100ms
    http_req_failed: ['rate<0.0001'],  // ошибок <0.01%
  },
};

let authToken;
let pvzId;
let hasActiveReception = false;

export function setup() {
  // 1. Получаем токен модератора
  const modLogin = http.post(
    'http://localhost:8080/dummyLogin',
    JSON.stringify({ role: 'moderator' }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  const modToken = modLogin.json('token');

  // 2. Создаем тестовый ПВЗ
  pvzId = uuidv4();
  http.post(
    'http://localhost:8080/pvz',
    JSON.stringify({
      id: pvzId,
      city: 'Москва',
      registrationDate: new Date().toISOString()
    }),
    { headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${modToken}`
    }}
  );

  // 3. Получаем токен сотрудника
  const empLogin = http.post(
    'http://localhost:8080/dummyLogin',
    JSON.stringify({ role: 'employee' }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  return {
    authToken: empLogin.json('token'),
    pvzId: pvzId
  };
}

export default function (data) {
  // 1. Создаем новую приемку (10% запросов)
  if (!hasActiveReception && Math.random() < 0.1) {
    const res = http.post(
      'http://localhost:8080/receptions',
      JSON.stringify({ pvzId: data.pvzId }),
      { headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${data.authToken}`
      }}
    );
    
    if (res.status === 201) {
      hasActiveReception = true;
    }
  }

  // 2. Пытаемся добавить товар
  const productTypes = ['электроника', 'одежда', 'обувь'];
  const productRes = http.post(
    'http://localhost:8080/products',
    JSON.stringify({
      type: productTypes[Math.floor(Math.random() * 3)],
      pvzId: data.pvzId
    }),
    { headers: { 
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${data.authToken}`
    }}
  );

  // 3. Проверяем результат
  if (hasActiveReception) {
    check(productRes, {
      'Товар добавлен в открытую приемку': (r) => r.status === 201,
      'Время ответа <100ms': (r) => r.timings.duration < 100
    });
  } else {
    check(productRes, {
      'Ошибка при добавлении в закрытую приемку': (r) => r.status === 400,
      'Сообщение об ошибке корректно': (r) => 
        r.json('message').includes('no active reception')
    });
  }

  // 4. Закрываем приемку (5% запросов)
  if (hasActiveReception && Math.random() < 0.05) {
    const closeRes = http.post(
      `http://localhost:8080/pvz/${data.pvzId}/close_last_reception`,
      null,
      { headers: { 'Authorization': `Bearer ${data.authToken}` }}
    );
    
    if (closeRes.status === 200) {
      hasActiveReception = false;
    }
  }
}