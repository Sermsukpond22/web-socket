/**
 * Tier 1: Feature Coverage - Authentication & Profile Tests
 */

const ApiClient = require('../lib/api_client');
const { subheader, pass, fail, info } = require('../lib/logger');

async function runTier1(ctx) {
  subheader('Tier 1: Feature Coverage (Auth & User Profile)');
  const api = new ApiClient(ctx.baseUrl);
  const testResults = [];

  // Define test credentials
  const userA = {
    username: 'user_a',
    email: 'user_a@test.com',
    password: 'password123'
  };

  const userB = {
    username: 'user_b',
    email: 'user_b@test.com',
    password: 'password123'
  };

  // Test 1.1: Register User A
  try {
    let resA = await api.register(userA.username, userA.email, userA.password);
    if (!resA.ok && (resA.status === 400 || resA.status === 409)) {
      // User might already exist from prior run, try login
      info(`User A (${userA.email}) already registered, logging in...`);
      resA = await api.login(userA.email, userA.password);
    }
    
    if (!resA.ok) {
      throw new Error(`Register/Login User A failed with status ${resA.status}: ${JSON.stringify(resA.data || resA.error)}`);
    }

    const tokenA = resA.data?.token || resA.data?.data?.token;
    const userAData = resA.data?.user || resA.data?.data?.user || resA.data?.data;
    
    if (!tokenA) {
      throw new Error(`JWT token missing in User A response: ${JSON.stringify(resA.data)}`);
    }

    ctx.userA = { ...userA, token: tokenA, id: userAData?.id || userAData?.ID };
    pass('Register / Retrieve User A (user_a@test.com)', `Token obtained, ID: ${ctx.userA.id || 'N/A'}`);
    testResults.push(true);
  } catch (err) {
    fail('Register / Retrieve User A (user_a@test.com)', err);
    testResults.push(false);
  }

  // Test 1.2: Register User B
  try {
    let resB = await api.register(userB.username, userB.email, userB.password);
    if (!resB.ok && (resB.status === 400 || resB.status === 409)) {
      info(`User B (${userB.email}) already registered, logging in...`);
      resB = await api.login(userB.email, userB.password);
    }

    if (!resB.ok) {
      throw new Error(`Register/Login User B failed with status ${resB.status}: ${JSON.stringify(resB.data || resB.error)}`);
    }

    const tokenB = resB.data?.token || resB.data?.data?.token;
    const userBData = resB.data?.user || resB.data?.data?.user || resB.data?.data;

    if (!tokenB) {
      throw new Error(`JWT token missing in User B response: ${JSON.stringify(resB.data)}`);
    }

    ctx.userB = { ...userB, token: tokenB, id: userBData?.id || userBData?.ID };
    pass('Register / Retrieve User B (user_b@test.com)', `Token obtained, ID: ${ctx.userB.id || 'N/A'}`);
    testResults.push(true);
  } catch (err) {
    fail('Register / Retrieve User B (user_b@test.com)', err);
    testResults.push(false);
  }

  // Test 1.3: Explicit Login as User A and User B
  try {
    const loginA = await api.login(userA.email, userA.password);
    if (!loginA.ok) {
      throw new Error(`Explicit login for User A failed with status ${loginA.status}`);
    }
    const tokenA = loginA.data?.token || loginA.data?.data?.token;
    if (tokenA && ctx.userA) ctx.userA.token = tokenA;

    const loginB = await api.login(userB.email, userB.password);
    if (!loginB.ok) {
      throw new Error(`Explicit login for User B failed with status ${loginB.status}`);
    }
    const tokenB = loginB.data?.token || loginB.data?.data?.token;
    if (tokenB && ctx.userB) ctx.userB.token = tokenB;

    pass('Login as User A and User B to verify authentication tokens');
    testResults.push(true);
  } catch (err) {
    fail('Login as User A and User B', err);
    testResults.push(false);
  }

  // Test 1.4: Verify GET /api/auth/me with Bearer token for User A
  try {
    if (!ctx.userA?.token) {
      throw new Error('User A token unavailable for /api/auth/me test');
    }
    const meA = await api.getMe(ctx.userA.token);
    if (!meA.ok) {
      throw new Error(`GET /api/auth/me failed for User A with status ${meA.status}`);
    }
    const profileA = meA.data?.user || meA.data?.data?.user || meA.data?.data || meA.data;
    const emailMatch = profileA?.email === userA.email;
    if (!emailMatch) {
      throw new Error(`User profile email mismatch: expected ${userA.email}, got ${profileA?.email}`);
    }
    if (!ctx.userA.id && (profileA?.id || profileA?.ID)) {
      ctx.userA.id = profileA.id || profileA.ID;
    }
    pass('Verify GET /api/auth/me for User A', `Email: ${profileA.email}`);
    testResults.push(true);
  } catch (err) {
    fail('Verify GET /api/auth/me for User A', err);
    testResults.push(false);
  }

  // Test 1.5: Verify GET /api/auth/me with Bearer token for User B
  try {
    if (!ctx.userB?.token) {
      throw new Error('User B token unavailable for /api/auth/me test');
    }
    const meB = await api.getMe(ctx.userB.token);
    if (!meB.ok) {
      throw new Error(`GET /api/auth/me failed for User B with status ${meB.status}`);
    }
    const profileB = meB.data?.user || meB.data?.data?.user || meB.data?.data || meB.data;
    const emailMatch = profileB?.email === userB.email;
    if (!emailMatch) {
      throw new Error(`User profile email mismatch: expected ${userB.email}, got ${profileB?.email}`);
    }
    if (!ctx.userB.id && (profileB?.id || profileB?.ID)) {
      ctx.userB.id = profileB.id || profileB.ID;
    }
    pass('Verify GET /api/auth/me for User B', `Email: ${profileB.email}`);
    testResults.push(true);
  } catch (err) {
    fail('Verify GET /api/auth/me for User B', err);
    testResults.push(false);
  }

  // Test 1.6: Verify GET /api/auth/me without token (Unauthenticated)
  try {
    const meNoToken = await api.getMe(null);
    if (meNoToken.status === 401 || meNoToken.status === 403 || (!meNoToken.ok && meNoToken.status !== 0)) {
      pass('GET /api/auth/me without token rejected correctly', `HTTP ${meNoToken.status}`);
      testResults.push(true);
    } else if (meNoToken.status === 0) {
      throw new Error('Server not reachable (network error status 0)');
    } else {
      throw new Error(`Expected 401 Unauthorized, got HTTP ${meNoToken.status}`);
    }
  } catch (err) {
    fail('GET /api/auth/me without token rejection', err);
    testResults.push(false);
  }

  return testResults.every(Boolean);
}

module.exports = runTier1;
