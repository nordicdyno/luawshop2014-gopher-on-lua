require "posix"
for i = 1, 10, 1 do
    print("Hello 🌝  & 🐻 " .. string.rep("!", i))
    posix.sleep(1)
end
