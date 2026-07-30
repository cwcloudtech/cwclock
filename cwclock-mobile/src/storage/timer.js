import AsyncStorage from "@react-native-async-storage/async-storage";

const TIMER_STORAGE_KEY = "cwclock.timer";

// Mirrors the CLI's timer.json (cwclock-cli/handlers/record.go's
// timerState) - persisted so backgrounding or killing the app doesn't lose
// an in-progress timer.
export const saveTimer = (timer) => AsyncStorage.setItem(TIMER_STORAGE_KEY, JSON.stringify(timer));

export const loadTimer = async () => {
  const raw = await AsyncStorage.getItem(TIMER_STORAGE_KEY);
  return raw ? JSON.parse(raw) : null;
};

export const clearTimer = () => AsyncStorage.removeItem(TIMER_STORAGE_KEY);
