#include <iostream>
#include <chrono>
#include <thread>

int main(){
  for (int i = 0; i < 1000; ++i) {
    std::cout << "hello, for the " << i << "th time" << std::endl;
    std::this_thread::sleep_for(std::chrono::milliseconds(1000));
  }
  return 0;
}
